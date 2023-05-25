package main

import (
	"fmt"
	"context"
	"strconv"
	"gorm.io/driver/postgres"
  "gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)
//GORM Product Model
type Product struct{
	Product_ID int `gorm:"primaryKey" gorm:"column:product_id"`
	Product_Name string `gorm:"column:product_name"`
	Product_Description string `gorm:"column:product_description"`
	Product_Images pq.StringArray `gorm:"type:text[]" gorm:"column:product_images"`
	Product_Price float64 `gorm:"column:product_price"`
	Compressed_Product_Images pq.StringArray `gorm:"type:text[]" gorm:"column:compressed_product_images"`
	Created_At time.Time `gorm:"column:created_at"`
	Updated_At time.Time `gorm:"column:updated_at"`
}
// GORM User Model 
type User struct{
	ID int `gorm:"column:id" gorm:"primaryKey"`
	Name string `gorm:"column:name"`
	Mobile int64 `gorm:"column:mobile"`
	Latitude float64 `gorm:"column:latitude"`
	Longitude float64 `gorm:"column:longitude"`
	Created_At time.Time `gorm:"column:created_at"`
	Updated_At time.Time `gorm:"column:updated_at"`
}
//Input JSON structure
type Input struct{
  User_ID   int `json:"user_id"`
  Product_Name  string `json:"product_name"`
  ProductDescription  string `json:"product_description"`
  ProductImages []string `json:"product_images"`
  ProductPrice  float64 `json:"product_price"`
}
func main(){
	//Database config
	dsn := "host=localhost user=postgres dbname=prod_db port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err!=nil{
		fmt.Println(err)
		panic(err)
	}else{
		fmt.Println("Database Connected Successfully!")
	}
	//RabbitMQ config
	conn,err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err!=nil{
		fmt.Println(err)
		panic(err)
	}
	defer conn.Close()
	//fmt.Println(conn)
	fmt.Println("Successfully connected to RabbitMQ!")
	
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db.AutoMigrate(&Product{}, &User{})
	fmt.Println("All schemas, tables, et cetera have been migrated")
  r := gin.Default()
  r.GET("/status", func(c *gin.Context){
    c.JSON(200, gin.H{"status":"Up and Running!",})
  })
  r.POST("/addProduct", func(c *gin.Context){
    var i Input
    err := c.BindJSON(&i);
    if err!=nil{
      c.JSON(400, gin.H{"error":err,})
    }
    prod := Product{Product_Name:i.Product_Name, Product_Description:i.ProductDescription, Product_Price:i.ProductPrice,Product_Images:i.ProductImages,}
    db.Save(&prod)
    c.JSON(200, "Product Added!")
    fmt.Println("Product ID:",prod.Product_ID)
    body := prod.Product_ID
		err = ch.PublishWithContext(
			ctx,
			"",
			q.Name,
			false,
			false,
			amqp.Publishing {
				ContentType: "text/plain",
				Body: []byte(strconv.Itoa(body)),
			},
		)
		if err!=nil{
			fmt.Println(err)
			panic(err)
		}
		fmt.Printf("[x] Sent %d\n", body)
  })
  r.Run(":5000")
}
