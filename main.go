package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/alteamc/minequery/v2"
	"github.com/caarlos0/env/v6"
	"github.com/zan8in/masscan"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var mongoCollection *mongo.Collection

type config struct {
	MongoURI        string `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDB         string `env:"MONGO_DB" envDefault:"minecraft"`
	MinecraftPort   int    `env:"MINECRAFT_PORT" envDefault:"25565"`
	ScanExcludeFile string `env:"SCAN_EXCLUDE_FILE" envDefault:"./exclude.conf"`
	ScanLimit       int    `env:"SCAN_LIMIT" envDefault:"100000"`
}

func pingMC(serverIP string, port int) {
	log.Printf("Pinging possible MC-Server %s:%d", serverIP, port)
	minequery.NewPinger(minequery.WithTimeout(10 * time.Second))
	res, err := minequery.Ping17(serverIP, port)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	samplePlayers := bson.A{}
	for _, player := range res.SamplePlayers {
		samplePlayers = append(samplePlayers, bson.M{"name": player.Nickname, "uuid": player.UUID})
	}
	mongoRes, err := mongoCollection.InsertOne(ctx, bson.D{
		{"ip", serverIP},
		{"version", res.VersionName},
		{"createdAt", time.Now()},
		{"motd", res.DescriptionText()},
		{"maxPlayers", res.MaxPlayers},
		{"onlinePlayers", res.OnlinePlayers},
		{"samplePlayers", samplePlayers},
	})
	if err != nil {
		return
	}
	log.Printf("Inserted new Server %s with ID %d\r\n", serverIP, mongoRes.InsertedID)
}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Panicf("%+v\n", err)
	}

	// --------------------------------------------------
	var exlucdedIPs []string
	file, err := os.Open(cfg.ScanExcludeFile)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		exlucdedIPs = append(exlucdedIPs, line)
	}
	if err := fileScanner.Err(); err != nil {
		log.Fatal(err)
	}

	// --------------------------------------------------

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	collection := client.Database(cfg.MongoDB).Collection("server")

	mongoCollection = collection

	// --------------------------------------------------

	var ipRanges []string
	for a := 0; a < 255; a++ {
		for b := 0; b < 255; b++ {
			possibleIPRange := fmt.Sprintf("%d.%d.0.0/16", a, b)
			ipRanges = append(ipRanges, possibleIPRange)
		}
	}

	// --------------------------------------------------

	for len(ipRanges) > 0 {
		rand.Seed(time.Now().Unix())
		randIdx := rand.Intn(len(ipRanges))
		selectedIPRange := ipRanges[randIdx]
		ipRanges = remove(ipRanges, randIdx)

		func() {
			ctx, cancel = context.WithTimeout(context.Background(), 1*time.Hour)
			defer cancel()
			scanPort := strconv.FormatInt(int64(cfg.MinecraftPort), 10)
			log.Printf("Started scanning range %s on Port %s", selectedIPRange, scanPort)
			scanner, err := masscan.NewScanner(
				masscan.SetParamTargets(selectedIPRange),
				masscan.SetParamPorts(scanPort),
				masscan.SetParamWait(10),
				masscan.SetParamRate(cfg.ScanLimit),
				masscan.SetParamExclude(exlucdedIPs...),
				masscan.WithContext(ctx),
			)

			if err != nil {
				log.Printf("unable to create masscan scanner: %v", err)
				return
			}

			if err := scanner.RunAsync(); err != nil {
				log.Print(err)
				return
			}

			stdout := scanner.GetStdout()
			stderr := scanner.GetStderr()

			go func() {
				for stdout.Scan() {
					srs := masscan.ParseResult(stdout.Bytes())
					atoi, err := strconv.Atoi(srs.Port)
					if err == nil {
						go pingMC(srs.IP, atoi)
					}
				}
			}()

			go func() {
				for stderr.Scan() {
					log.Println("err: ", stderr.Text())
				}
			}()

			if err := scanner.Wait(); err != nil {
				log.Println(err)
			}
			time.Sleep(15 * time.Second)
		}()
	}

	log.Println("Finished Scanning")
}
