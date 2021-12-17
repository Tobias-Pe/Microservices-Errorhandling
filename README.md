
## Badges

[![MIT License](https://img.shields.io/apm/l/atomic-design-ui.svg?)](https://github.com/tterb/atomic-design-ui/blob/master/LICENSEs)

[![Gitmoji](https://img.shields.io/badge/gitmoji-%20😜%20😍-FFDD67.svg)](https://gitmoji.dev)



# Errorhandling in a Microservice Architecture

The application is a Demo-Webshop, where 9 Services communicate using synchronous and asynchronous communication. 

The aim of the application is to implement various error handling methods, try them out and measure their effectiveness under pressure of load tests.

_This project is part of a Bachelorthesis in Computer Science🎓_


## API Reference

#### Get all articles 🛍️

```http
  GET /articles/${category}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `category`      | `string` | **Optional**. Filter the articles for a certain category. |

#### Get an exchange rate 💰

```http
  GET /exchange/${currency}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `currency`      | `string` | **Required**. Currently supported: USD,GBP,INR,CAS,JPY,SEK,PLN |

Mocks an exchange from ${currency} to €

#### Create a shopping cart 🛒

```http
  POST /cart
```

| JSON-Body | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `article_id`      | `string` | **Optional**. Will create a cart with this article_id already in it |

Returns the new carts ID.

#### Get a cart's content 🛒

```http
  GET /cart/${id}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `id`      | `string` | **Required**. The ID of your created cart. |

#### Create an order 🧾

```http
  POST /order
```

| JSON-Body | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `cartId`      | `string` | **Required**  |
| `name`      | `string` | **Required**  |
| `address`      | `string` | **Required**  |
| `creditCard`      | `string` | **Required**  |

Creates an order, that will be validated and shipped in the background.

#### Get an order's content 🧾

```http
  GET /order/${id}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `id`      | `string` | **Required**. The ID of your created order. |

Look at the current status of your order.

## Tech Stack

**Order & Stock:** MongoDB with ACID-Transactions

**Cart:** Redis

**API-Gateway:** gin-gonic/gin

**Synchronous Communication:** GRPC

**Asynchronous Communication:** RabbitMQ

**Load Balancing:** NGINX with Docker DNS
## Run Locally

Clone the project

```bash
  git clone https://link-to-project
```

Go to the project directory

```bash
  cd my-project
```

Start all containers with compose

```bash
  docker-compose up --force-recreate --build -V --remove-orphans
```

_PS: there are some run configurations in the .run folder_
## Deployment

With docker-swarm TBD


## Authors

[@tobiaspeslalz](https://github.com/Tobias-Pe)

[![linkedin](https://img.shields.io/badge/linkedin-0A66C2?style=for-the-badge&logo=linkedin&logoColor=white)](www.linkedin.com/in/tobias-peslalz)
