# Zocket Golang Assignment
#### Name: Mukund Deepak

## Note:
Hi there! Please note that I have no experience writing testing reports but this pretty much covers all the functionality and gives you an understanding of the Application.

I have also encapsulated everything into a monolith which is not a good practice, but i assure you, in production, I will use modular programming and microservice architecture to decouple and make a clean codebase.

## Problem Statement:
I have created as per request, an API which has functionality to add products to the DB and then sends the product\_id to the Rabbit MQ message queue initialized from which the consumer who is always listening for new messages in the queue, recieves the PID, get's the image\_urls of that particular product from the DB and then downloads each of the image and adds the compressed equivalent to the local directory and adds the path to the compressed image path array in the DB for that particular product. This application is a very basic working model, therefore it only works with jpegs as of now.

## Code:
**app.go (This is the API):**
```go 
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
type User struct{
	ID int `gorm:"column:id" gorm:"primaryKey"`
	Name string `gorm:"column:name"`
	Mobile int64 `gorm:"column:mobile"`
	Latitude float64 `gorm:"column:latitude"`
	Longitude float64 `gorm:"column:latitude"`
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
    //fmt.Printf("%T\n", i.ProductImages)
    prod := Product{Product_Name:i.Product_Name, Product_Description:i.ProductDescription, Product_Price:i.ProductPrice,Product_Images:i.ProductImages,}
    db.Save(&prod)
    c.JSON(200, "Product Added!")
    fmt.Println("Product ID:",prod.Product_ID)
    //fmt.Printf("%T\n", prod.Product_Images)
    //For testing purposes, i used the below code.
    /*fmt.Printf("User ID:%d\nProduct Name:%s\nProduct Description:%s\nProduct Price:%f\n", i.User_ID, i.Product_Name, i.ProductDescription, i.ProductPrice);
    for _,value := range i.ProductImages{
      fmt.Printf("%s\n",value)
    }*/
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
```

**consumer.go:**
```go
package main

import(
  "fmt"
  "log"
  "strconv"
  "strings"
  "os"
  "image"
  "image/jpeg"
  "time"
  "net/http"
  "gorm.io/driver/postgres"
  "gorm.io/gorm"
  "github.com/lib/pq"
  "github.com/nfnt/resize"
  amqp "github.com/rabbitmq/amqp091-go"
)
func extractImageName(url string) string {
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]
	return strings.TrimSuffix(filename, ".jpg")
}
func DownloadAndCompress(urls []string) ([]string){
  var final_arr []string 
  for _, url := range urls{
    response, err := http.Get(url)
    if err!=nil{
      fmt.Println("Failed to retrieve image!")
    }
    defer response.Body.Close()
    img_name := extractImageName(url)
    _, err = os.Stat("compressed_images")
    if os.IsNotExist(err){
      err:=os.Mkdir("compressed_images", 0755)
      if err!=nil{
        fmt.Println(err)
        panic(err)
      }
    }
    file, err := os.Create("compressed_images/"+img_name+"_compressed.jpg")
    if err!=nil{
      fmt.Println(err)
      panic(err)
    }
    defer file.Close()
    var img image.Image
    img, err = jpeg.Decode(response.Body)
    if err!=nil{
      fmt.Println(err)
      panic(err)
    }
    newImg := resize.Thumbnail(200,200,img,resize.Lanczos3)
    err = jpeg.Encode(file, newImg, nil)
    if err!=nil{
      fmt.Println(err)
      panic(err)
    }

    final_arr = append(final_arr, file.Name())
  }
  return final_arr
}
//Product Model for Retrieval
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
func main(){
  dsn:="host=localhost user=postgres dbname=prod_db port=5432 sslmode=disable TimeZone=UTC"
  db, err := gorm.Open(postgres.Open(dsn))
  if err!=nil{
    fmt.Println(err)
    panic(err)
  }else{
    fmt.Println("Successfully connected to Postgres DB")
  }

  //RabbitMQ connection
  conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
  defer conn.Close()
  if err!=nil{
    fmt.Println(err)
    panic(err)
  }else{
    fmt.Println("RabbitMQ connection Successful!")
  }
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
      log.Printf("Recieved a PID: %s\n", d.Body);
      var prod Product
      id, err := strconv.Atoi(string(d.Body))
      if err!=nil{
        fmt.Println(err)
        panic(err)
      }
      db.Where("product_id = ?", id).First(&prod)
      compressed_li := DownloadAndCompress([]string(prod.Product_Images))
      prod.Compressed_Product_Images = pq.StringArray(compressed_li)
      db.Save(&prod)

      fmt.Println(prod)
    }
  }()

  fmt.Println("[*] Waiting for Messages from queue... To exit press Ctrl+C")
  <-forever
}
```

## Testing:

### Unit Testing:
#### API:
API requirements:
Design an API where it should receive a product data and store in the database, below
are the parameters that should be passed in the API
- user\_id (create users table and primary key of that table)
- product\_name
- product\_description (text)
- product\_images (array of image urls)
- product\_price (Number)

The API has two endpoints as of now:
/addProduct (Used for adding the products to DB)
/status (Used for checking running status of API)

**/status endpoint:**
![](./testing_images/status.png)

**/addProduct endpoint:**
The addProduct endpoint according to API Requirements need to add product details to DB aka populate it. But in this aspect, product also acts as a **producer**, so it should add the product ID to the queue.

Populating the DB:    
![](./testing_images/addProduct.png)

DB Verification:    
![](./testing_images/addProduct2.png)

If table does not exists before, as soon as app.go is ran, it takes care of creation and migration of models i.e. creating the tables.

### Producer:
As mentioned earlier, API acts as producer and adds the newly created product's id into the queue.

**Adding product id to products queue:**
A queue called products is created under a channel which can be seen in the RabbitMQ Management Portal.

Adding the product to DB and pushing product id to message queue:    
![](./testing_images/addProduct.png)

![](./testing_images/producer_terminal.png)

Let's verify this in the RabbitMQ management portal:    
![](./testing_images/producer_verification_1.png)

![](./testing_images/producer_verification_2.png)

### Consumer:
The consumer keeps listening for new messages in the products queue, retrieves the product id and obtains the image url array and passes it to the DownloadAndCompress function which retrieves all the images, compresses it and saves it in a compressed jpeg file locally inside the compressed\_images folder and adds the path to the compressed\_product\_images array which is updated to the DB at the end.

**Consumer pulling a product_id and printing the array of compressed_product_images after saving in local:**
![](./testing_images/consumer_1.png)

**Saved in compressed_images folder (Verification):**
![](./testing_images/consumer_2.png)

**Size Comparision:**
We will take the red kitten picture to compare.

Before compression:    
![](./testing_images/consumer_3.png)

After compression:    
![](./testing_images/consumer_4.png)

**Updated array of compressed image path in DB:**
![](./testing_images/consumer_5.png)


### Populating User table manually in psql:
**Inserting values into table users:**
![](./testing_images/user1.png)

**Verification of Addition of records into table users:**
![](./testing_images/user2.png)

## Integration Testing:
Ok so we shall add the following data to the Products DB through the API:
```json
{
    "user_id": 2,
    "product_name": "Cello Butterfree",
    "product_description": "Ball Point Pen",
    "product_images": ["https://m.media-amazon.com/images/I/31hFt5mqAoL._SY355_.jpg", "https://m.media-amazon.com/images/I/318yfV6jnIL._SY355_.jpg", "https://assetscdn1.paytm.com/images/catalog/product/S/ST/STACELLO-BUTTERHEAD8922516E054A9/1563558363956_2..jpg"],
    "product_price": 70
}
```

**Postman Request:**
![](./testing_images/int1.png)

We successfully send the postman request and get the response as follows:
![](./testing_images/int2.png)

The terminal output confirms a successful request and addition of product ID to queue.
![](./testing_images/int3.png)

Parallely, our consumer who has always been listening should have recieved the same PID and for verification printed out the DB entry after compressing all images and placing it in it's respective path and updating the DB with the array of compressed image paths.
![](./testing_images/int4.png)

In the compressed directory, we should see three images of cello butterfree pen resized and compressed.
![](./testing_images/int5.png)

There is a decrease in size as seen in unit testing for consumer.

## Benchmark Testing:
I'm quite unfamiliar with performance testing and benchmark testing so I am generalizing performance observations.

I am making use of one RabbitMQ queue which can handle more than 20K+ messages per second. The API has little to no performance delay since it is running in local and has very few handlers. We can see a downgrade in performance if the API was hosted online. Due to networks delays, et cetera.
