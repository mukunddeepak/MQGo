package main

import (
	"fmt"
	"gorm.io/driver/postgres"
  "gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
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
	if err != nil && db==nil{
		fmt.Println(err)
	}else{
		fmt.Println("Database Connection Successful!!!")
	}
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
    fmt.Printf("%T\n", i.ProductImages)
    prod := Product{Product_Name:i.Product_Name, Product_Description:i.ProductDescription, Product_Price:i.ProductPrice,Product_Images:i.ProductImages,}
    db.Save(&prod)
    fmt.Println("Product ID:",prod.Product_ID)
    fmt.Printf("%T\n", prod.Product_Images)
    //For testing purposes, i used the below code.
    /*fmt.Printf("User ID:%d\nProduct Name:%s\nProduct Description:%s\nProduct Price:%f\n", i.User_ID, i.Product_Name, i.ProductDescription, i.ProductPrice);
    for _,value := range i.ProductImages{
      fmt.Printf("%s\n",value)
    }*/

  })
  r.Run(":5000")
}