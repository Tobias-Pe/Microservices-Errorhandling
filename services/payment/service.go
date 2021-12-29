/*
 * MIT License
 *
 * Copyright (c) 2021 Tobias Leonhard Joschka Peslalz
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package payment

import (
	"encoding/json"
	"fmt"
	"github.com/Tobias-Pe/Microservices-Errorhandling/api/requests"
	loggingUtil "github.com/Tobias-Pe/Microservices-Errorhandling/pkg/log"
	"github.com/Tobias-Pe/Microservices-Errorhandling/pkg/metrics"
	"github.com/Tobias-Pe/Microservices-Errorhandling/pkg/models"
	loggrus "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"math"
	"math/rand"
	"strings"
	"time"
)

const (
	methodListenOrders = "PayOrder"
	methodPublishOrder = "PublishOrder"
)

var logger = loggingUtil.InitLogger()

type Service struct {
	AmqpChannel    *amqp.Channel
	AmqpConn       *amqp.Connection
	orderMessages  <-chan amqp.Delivery
	rabbitUrl      string
	requestsMetric *metrics.RequestsMetric
}

func NewService(rabbitAddress string, rabbitPort string) *Service {
	service := &Service{}
	service.rabbitUrl = fmt.Sprintf("amqp://guest:guest@%s:%s/", rabbitAddress, rabbitPort)
	var err error = nil
	// retry connecting to rabbitmq
	for i := 0; i < 6; i++ {
		err = service.initAmqpConnection()
		if err == nil {
			break
		}
		logger.Infof("Retrying... (%d/%d)", i, 5)
		time.Sleep(time.Duration(int64(math.Pow(2, float64(i)))) * time.Second)
	}
	if err != nil {
		return nil
	}

	err = service.createOrderListener()
	if err != nil {
		return nil
	}

	service.requestsMetric = metrics.NewRequestsMetrics()

	return service
}

func (service *Service) initAmqpConnection() error {
	conn, err := amqp.Dial(service.rabbitUrl)
	if err != nil {
		return err
	}
	// connection and channel will be closed in main
	service.AmqpConn = conn
	service.AmqpChannel, err = conn.Channel()
	if err != nil {
		return err
	}
	// prefetchCount 1 in QoS will load-balance messages between many instances of this service
	err = service.AmqpChannel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}
	return nil
}

// createOrderListener initialises exchange and queue and binds the queue to a topic and routing key to listen from
func (service *Service) createOrderListener() error {
	err := service.AmqpChannel.ExchangeDeclare(
		requests.OrderTopic, // name
		"topic",             // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return err
	}
	q, err := service.AmqpChannel.QueueDeclare(
		"payment_"+requests.OrderTopic+"_queue", // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}
	err = service.AmqpChannel.QueueBind(
		q.Name,                  // queue name
		requests.OrderStatusPay, // routing key
		requests.OrderTopic,     // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// orderMessages will be where we get our order messages from
	service.orderMessages, err = service.AmqpChannel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}
	return nil
}

// mockPayment simulates the paying process. Validates creditCart string and then sleeps random to simulate working
func (service *Service) mockPayment(creditCard string) bool {
	creditCard = strings.ToLower(strings.TrimSpace(creditCard))
	// simulate work: sleep for every char in creditCard randomly
	for _, charVariable := range creditCard {
		timeout := rand.Intn(5)
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		// validate for only numbers and special symbols
		if charVariable >= 'a' && charVariable <= 'z' {
			return false
		}
	}
	return true
}

// mockPaymentRollback simulates undoing the payment process
func (service *Service) mockPaymentRollback(creditCard string) {
	creditCard = strings.ToLower(strings.TrimSpace(creditCard))
	// simulate work: sleep for every char in creditCard randomly
	for _, charVariable := range creditCard {
		timeout := rand.Intn(5)
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		if charVariable >= 'a' && charVariable <= 'z' {
			return
		}
	}
}

// ListenOrders reads out order messages from bound amqp queue
func (service *Service) ListenOrders() {
	for message := range service.orderMessages {
		order := &models.Order{}
		err := json.Unmarshal(message.Body, order)
		if err == nil {
			isAllowed := service.mockPayment(order.CustomerCreditCard)
			err = message.Ack(false)
			service.requestsMetric.Increment(err, methodListenOrders)
			if err != nil {
				logger.WithError(err).Error("Could not ack message.")
				if isAllowed { // ack could not be sent but transaction was successfully
					logger.WithFields(loggrus.Fields{"request": *order}).WithError(err).Info("Rolling back transaction...")
					// rollback transaction. because of the missing ack the current request will be resent
					service.mockPaymentRollback(order.CustomerCreditCard)
					logger.WithFields(loggrus.Fields{"request": *order}).Info("Rolling back successfully")
				}
			} else {
				if !isAllowed { // abort order because of invalid payment data
					logger.WithFields(loggrus.Fields{"payment_status": isAllowed, "request": *order}).Warn("Payment unsuccessfully. Aborting order...")
					status := models.StatusAborted("We could not get the needed amount from your credit card. Please check your account.")
					order.Status = status.Name
					order.Message = status.Message
				} else {
					logger.WithFields(loggrus.Fields{"request": *order}).Infof("Order payed.")
					status := models.StatusShipping()
					order.Status = status.Name
					order.Message = status.Message
				}
				// broadcast updated order
				err = order.PublishOrderStatusUpdate(service.AmqpChannel)
				if err != nil {
					logger.WithFields(loggrus.Fields{"request": *order}).WithError(err).Error("Could not publish order update")
				}
				service.requestsMetric.Increment(err, methodPublishOrder)
			}
		} else {
			logger.WithError(err).Error("Could not unmarshall message")
			// ack message despite the error, or else we will get this message repeatedly
			err = message.Ack(false)
			if err != nil {
				logger.WithError(err).Error("Could not ack message.")
			}
			service.requestsMetric.Increment(err, methodListenOrders)
		}
	}
	logger.Warn("Stopped Listening for Orders! Restarting...")
	// try reconnecting
	err := service.createOrderListener()
	if err != nil {
		logger.WithError(err).Error("Stopped Listening for Orders! Could not restart")
	} else {
		service.ListenOrders()
	}
}
