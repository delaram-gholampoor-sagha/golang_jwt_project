package controllers

import (
	"context"
	"fmt"
	helper "golang_jwt_project/helpers"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Delaram-Gholampoor-Sagha/golang_jwt_project/database"
	helper "github.com/Delaram-Gholampoor-Sagha/golang_jwt_project/helpers"
	"github.com/Delaram-Gholampoor-Sagha/golang_jwt_project/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

//  "go.mongodb.org/mongo-driver/mongo"

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

// we cant store our password just as it is we have to hash it 
func HashPassword(password string) string  {
	bcrypt.GenerateFromPassword([]byte(password) , 14)
	if err != nil {
          log.Panic(err)
	}
	return string(bytes)
}


func VerifyPassword(userPasswordString string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(userPasswordString), []byte(providedPassword))

	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}
	return check, msg


}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationError := validate.Struct(user)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email "})
		}


		password := HashPassword(*user.Password)
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone"})
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this phone or email is already being used"})
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshTOken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_Type, *user.User_id)
		user.Token = &token
		user.Refresh_Token = &refreshTOken
		resultInserstionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("user item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInserstionNumber)
	}

}

func LogIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, candel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		var foundUser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer candel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": " email or password is incorrect"})
			return
		}
		passwordISValid, msg := VerifyPassword(*user.Password, &foundUser.Password)
		defer candel()

		if passwordISValid != true {
			c.JSON(http.StatusInternalServerError , gin.H{"error" : msg})
		}

		if foundUser.Email == nill {
			c.JSON(http.StatusInternalServerError , gin.H{"error" : "user not found"})
		}
		token , refreshToken := helper.GenerateAllTokens(*foundUser.Email , * foundUser.First_name , *foundUser.Last_name , *foundUser.User_id  , foundUser.User_Type ) 
		helper.UpdateAllTokens(token , refreshToken , foundUser.User_id )
	          err := userCollection.FindOne(ctx , bson.M{"user_id" : foundUser.User_id}).Decode(&foundUser)
	  if err != nil {
		  c.JSON(http.StatusInternalServerError , gin.H{"error" : err.Error()})
		  return
	  } 

	  c.JSON(http.StatusOK , foundUser)
	}
}

// this function could only be availbe to admin
func GetUsers() gin.HandlerFunc{
    return func(c *gin.Context) {
		helper.CheckUserType(c , "ADMIN"); err != nil {
			 c.JSON(http.StatusBadRequest , gin.H{"error" : err.Error} )
			 return 
		} 
		var ctx , cancel := context.WithTimeout(context.Background() , 100*time.Now())
		recordPerPage , err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10 
		}
		page , err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page <1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex , err := strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.M{{"%match" , bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}}, 
			{"total_count", bson.D{{"$sum", 1}}}, 
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
			result,err := userCollection.Aggregate(ctx, mongo.Pipeline{
				matchStage, groupStage, projectStage})
			defer cancel()
			if err!=nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
			}
			var allusers []bson.M
			if err = result.All(ctx, &allusers); err!=nil{
				log.Fatal(err)
			}
			c.JSON(http.StatusOK, allusers[0])
	}
}

// only the admin can see the others user data
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		// we need this function because we want to check if the user is admin or not
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)

	}
}
