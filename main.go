package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Student struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string             `bson:"name" json:"name"`
	Email string             `bson:"email" json:"email"`
}

var client *mongo.Client
var studentCollection *mongo.Collection

func main() {
	// Set up MongoDB client
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Set up MongoDB client
	var err error
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		panic(err)
	}

	studentCollection = client.Database("school").Collection("students")
	router := gin.Default()

	// Middleware to log requests
	router.Use(gin.Logger())

	// Versioned routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/students", addStudent)
		v1.GET("/students", getAllStudents)
		v1.GET("/students/:id", getStudentByID)
		v1.PUT("/students/:id", updateStudent)
		v1.DELETE("/students/:id", deleteStudent)

		// Health check endpoint
		v1.GET("/healthcheck", healthCheck)
	}

	router.Run("0.0.0.0:3000")
}

func addStudent(c *gin.Context) {
	var student Student
	if err := c.ShouldBindJSON(&student); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student.ID = primitive.NewObjectID()

	_, err := studentCollection.InsertOne(context.TODO(), student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, student)
}

func getAllStudents(c *gin.Context) {
	var students []Student
	cursor, err := studentCollection.Find(context.TODO(), bson.D{{}})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var student Student
		cursor.Decode(&student)
		students = append(students, student)
	}

	c.JSON(http.StatusOK, students)
}

func getStudentByID(c *gin.Context) {
	id := c.Param("id")

	objID, _ := primitive.ObjectIDFromHex(id)

	var student Student
	err := studentCollection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&student)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	c.JSON(http.StatusOK, student)
}

func updateStudent(c *gin.Context) {
	id := c.Param("id")

	objID, _ := primitive.ObjectIDFromHex(id)

	var updatedData Student
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := studentCollection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": updatedData})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedData)
}

func deleteStudent(c *gin.Context) {
	id := c.Param("id")

	objID, _ := primitive.ObjectIDFromHex(id)

	_, err := studentCollection.DeleteOne(context.TODO(), bson.M{"_id": objID})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}
