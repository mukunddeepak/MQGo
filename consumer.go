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
