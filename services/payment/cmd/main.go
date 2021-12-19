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

package main

import (
	loggingUtil "github.com/Tobias-Pe/Microservices-Errorhandling/pkg/log"
	"github.com/Tobias-Pe/Microservices-Errorhandling/services/payment"
	loggrus "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type configuration struct {
	rabbitAddress string
	rabbitPort    string
}

var logger = loggingUtil.InitLogger()

func main() {
	config := readConfig()
	createServer(config)
}

func createServer(configuration configuration) {
	service := payment.NewService(
		configuration.rabbitAddress,
		configuration.rabbitPort,
	)

	logger.Infof("Server listening...")
	service.ListenOrders()

	err := service.AmqpChannel.Close()
	if err != nil {
		logger.WithError(err).Error("Error on closing amqp-channel")
	}

	err = service.AmqpConn.Close()
	if err != nil {
		logger.WithError(err).Error("Error on closing amqp-connection")
	}
}

func readConfig() configuration {
	viper.SetConfigType("env")
	viper.SetConfigName("local")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		logger.Info(err)
	}

	rabbitAddress := viper.GetString("RABBIT_MQ_ADDRESS")
	rabbitPort := viper.GetString("RABBIT_MQ_PORT")

	logger.WithFields(loggrus.Fields{
		"RABBIT_MQ_ADDRESS": rabbitAddress,
		"RABBIT_MQ_PORT":    rabbitPort,
	}).Info("config variables read")

	return configuration{
		rabbitAddress: rabbitAddress,
		rabbitPort:    rabbitPort,
	}
}
