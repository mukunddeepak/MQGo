package main

import(
  "fmt"
  "log"
  amqp "github.com/rabbitmq/amqp091-go"
)

func main(){
  conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
  defer conn.Close()

  ch,err := conn.Channel()
  if err!=nil{
    fmt.Println(err)
    panic(err)
  }
  defer ch.Close()

  q, err := ch.QueueDeclare(
    "product",
    false,
    false,
    false,
    false,
    nil,
  )
  if err!=nil{
    fmt.Println(err)
    panic(err)
  }
  msgs, err := ch.Consume(
    q.Name,
    "",
    true,
    false,
    false,
    false,
    nil, 
  )
  if err!=nil{
    fmt.Println(err)
    panic(err)
  }

  var forever chan struct{}

  go func(){
    for d:=range msgs{
      log.Printf("Recieved a message: %s\n", d.Body);
    }
  }()

  fmt.Println("[*] Waiting for Messages from queue... To exit press Ctrl+C")
  <-forever
}
