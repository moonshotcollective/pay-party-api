package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hansmrtn/pay-party-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// Get env variables
var mongoURI = os.Getenv("DATABASE_URL")
var dbCollection = os.Getenv("DATABASE_COLLECTION")
var dbName = os.Getenv("DATABASE_NAME")
var port = os.Getenv("PORT")

var mg MongoInstance

func init() {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := context.WithTimeout(context.Background(), 30*time.Second)
	defer stop()
	err = client.Connect(ctx)
	db := client.Database(dbName)
	if err != nil {
		log.Fatal(err)
	}
	mg = MongoInstance{
		Client: client,
		DB:     db,
	}
}

// Get all parties
func GetAllParties(ctx *fiber.Ctx) error {
	log.Println("GET /parties")
	query := bson.D{{}}
	cursor, err := mg.DB.Collection(dbCollection).Find(ctx.Context(), query)
	if err != nil {
		log.Fatal(err.Error())
		return ctx.Status(500).SendString(err.Error())
	}
	var parties []models.Party = make([]models.Party, 0)
	// iterate the cursor and decode each item into a party
	if err := cursor.All(ctx.Context(), &parties); err != nil {
		return ctx.Status(500).SendString(err.Error())
	}
	return ctx.JSON(parties)
}

// Get a party
func GetParty(ctx *fiber.Ctx) error {
	partyID, err := primitive.ObjectIDFromHex(
		ctx.Params("id"),
	)
	if err != nil {
		return ctx.SendStatus(400)
	}
	party := new(models.Party)
	query := bson.D{{Key: "_id", Value: partyID}}
	err = mg.DB.Collection(dbCollection).FindOne(ctx.Context(), query).Decode(&party)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return ctx.SendStatus(404)
		}
		return ctx.SendStatus(500)
	}
	return ctx.Status(200).JSON(party)
}

// Create a new party
func NewParty(ctx *fiber.Ctx) error {
	collection := mg.DB.Collection(dbCollection)
	party := new(models.Party)
	if err := ctx.BodyParser(party); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}
	// ensure mongo always sets generated ObjectIDs
	party.ID = ""
	insertRes, err := collection.InsertOne(ctx.Context(), party)
	if err != nil {
		return ctx.Status(500).SendString(err.Error())
	}
	filter := bson.D{{Key: "_id", Value: insertRes.InsertedID}}
	createdRecord := collection.FindOne(ctx.Context(), filter)
	createdParty := &models.Party{}
	createdRecord.Decode(createdParty)
	return ctx.Status(200).JSON(createdParty)
}

// Update a party
func UpdateParty(ctx *fiber.Ctx) error {
	partyID, err := primitive.ObjectIDFromHex(
		ctx.Params("id"),
	)
	if err != nil {
		return ctx.SendStatus(400)
	}
	party := new(models.Party)
	if err := ctx.BodyParser(party); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}
	query := bson.D{{Key: "_id", Value: partyID}}
	update := bson.D{
		{Key: "$set",
			Value: bson.D{
				{Key: "ballots", Value: party.Ballots},
			},
		},
	}
	err = mg.DB.Collection(dbCollection).FindOneAndUpdate(ctx.Context(), query, update).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ctx.SendStatus(404)
		}
		return ctx.SendStatus(500)
	}
	// return updated ObjectId
	party.ID = ctx.Params("id")
	return ctx.Status(200).JSON(party)
}

// Delete party
// Docs: https://docs.mongodb.com/manual/reference/command/delete/
func DeleteParty(ctx *fiber.Ctx) error {
	partyID, err := primitive.ObjectIDFromHex(
		ctx.Params("id"),
	)
	if err != nil {
		return ctx.SendStatus(400)
	}
	// find and delete the party with the given ID
	query := bson.D{{Key: "_id", Value: partyID}}
	result, err := mg.DB.Collection(dbCollection).DeleteOne(ctx.Context(), &query)
	if err != nil {
		return ctx.SendStatus(500)
	}
	if result.DeletedCount < 1 {
		return ctx.SendStatus(404)
	}
	// the record was deleted
	return ctx.SendStatus(204)
}

func main() {
	// Create Fiber App
	app := fiber.New()
	// Use CORS
	// TODO: Configure CORS
	app.Use(cors.New())
	// App routes
	app.Get("/parties", GetAllParties)
	app.Get("/party/:id", GetParty)
	app.Post("/party", NewParty)
	app.Put("/party/:id", UpdateParty)
	app.Delete("/party/:id", DeleteParty)
	log.Fatal(app.Listen(":" + port))
}
