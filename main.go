package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	DB     *mongo.Database
}

type Data struct {
	Party  string `json:"party"`
	Ballot struct {
		Address string `json:"address"`
		Votes   string `json:"votes"`
	} `json:"ballot"`
}
type Ballot struct {
	Signature string `json:"signature"`
	Data      Data   `json:"data"`
}

type Config struct {
	Strategy string `json:"strategy"`
	Nvotes   int32  `json:"nvotes"`
}

type Receipt struct {
	Account string      `json:"account"`
	Amount  interface{} `json:"amount"` // BigNumbers
	Token   string      `json:"token"`
	Txn     string      `json:"txn"`
}
type Party struct {
	ID           string    `json:"id,omitempty" bson:"_id,omitempty"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Config       Config    `json:"config"`
	Receipts     []Receipt `json:"receipts"` // Yo
	Participants []string  `json:"participants"`
	Candidates   []string  `json:"candidates"`
	Ballots      []Ballot  `json:"ballots"`
}

var mg MongoInstance

var mongoURI = "mongodb+srv://doapps-148ccc10-e8d4-4716-946e-bb2a5ba60dd5:67UQ8cf9rJ130VL5@db-mongodb-sfo3-pay-party-30d250ca.mongo.ondigitalocean.com/admin?authSource=admin&tls=true" //os.Getenv("DATABASE_URL")
var dbName = os.Getenv("DATABASE_NAME")
var dbHost = os.Getenv("DATABASE_HOST")
var dbCollection = os.Getenv("DATABASE_COLLECTION")
var dbPort = os.Getenv("DATABASE_PORT")
var dbUser = os.Getenv("DATABASE_USERNAME")
var dbPass = os.Getenv("DATABASE_PASSWORD")
var dbParams = os.Getenv("DATABASE_PARAMS")
var dbCert = os.Getenv("CA_CERT")
var port = os.Getenv("PORT")

func Connect() error {

	URI := mongoURI + "&tlsInsecure=true" // + "&tlsCAFile=" + dbCert
	// log.Printf("Connecting to URI: %s", URI)

	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		return err
	}

	ctx, stop := context.WithTimeout(context.Background(), 30*time.Second)
	defer stop()

	err = client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil {
		return err
	}

	mg = MongoInstance{
		Client: client,
		DB:     db,
	}

	// log.Printf("Connected to DB URL: %s\nOn DB Name: %s\nCollection Name: %s\nPort: %s", mongoURI, dbName, dbCollection, port)

	return nil
}

func main() {

	log.Printf("mongodb+srv://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbParams)

	// Connect to the db
	if err := Connect(); err != nil {
		log.Fatal(err)
	}

	// Create Fiber App
	app := fiber.New()

	app.Use(cors.New())

	// get all parties from the db
	app.Get("/parties", func(ctx *fiber.Ctx) error {
		log.Println("GET /parties")
		query := bson.D{{}}
		cursor, err := mg.DB.Collection(dbCollection).Find(ctx.Context(), query)
		if err != nil {
			log.Fatal(err.Error())
			return ctx.Status(500).SendString(err.Error())
		}
		var parties []Party = make([]Party, 0)
		// iterate the cursor and decode each item into a party
		if err := cursor.All(ctx.Context(), &parties); err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		return ctx.JSON(parties)
	})

	// get a party with ObjectId from db
	app.Get("/party/:id", func(ctx *fiber.Ctx) error {
		partyID, err := primitive.ObjectIDFromHex(
			ctx.Params("id"),
		)
		if err != nil {
			return ctx.SendStatus(400)
		}
		party := new(Party)
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
	})

	// Create a party and insert into db
	app.Post("/party", func(ctx *fiber.Ctx) error {
		collection := mg.DB.Collection(dbCollection)

		party := new(Party)

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
		createdParty := &Party{}
		createdRecord.Decode(createdParty)

		return ctx.Status(200).JSON(createdParty)

	})

	app.Put("/party/:id", func(ctx *fiber.Ctx) error {
		partyID, err := primitive.ObjectIDFromHex(
			ctx.Params("id"),
		)
		if err != nil {
			return ctx.SendStatus(400)
		}
		party := new(Party)
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

	})

	// Delete party
	// Docs: https://docs.mongodb.com/manual/reference/command/delete/
	app.Delete("/party/:id", func(ctx *fiber.Ctx) error {
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
	})

	log.Fatal(app.Listen(":8080"))

}
