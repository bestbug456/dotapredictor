package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	rprop "github.com/bestbug456/gorpropplus"
	gorest "github.com/fredmaggiowski/gorest"

	"gopkg.in/mgo.v2"
)

var heroID map[string]int

func init() {
	heroID = make(map[string]int)
	for i := 0; i < len(heros); i++ {
		heroID[heros[i]] = i
	}
}

func main() {

	address := os.Getenv("address")
	username := os.Getenv("username")
	password := os.Getenv("password")
	option := os.Getenv("option")
	ssl := os.Getenv("ssl")

	var s *mgo.Session
	var err error
	log.Printf("Accessing to db\n")
	if ssl == "false" {
		log.Printf("Accessing to db via mgo.Dial\n")
		s, err = mgo.Dial(fmt.Sprintf("mongodb://%s:%s@%s/%s", username, password, address, option))
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			os.Exit(1)
		}
	} else {
		log.Printf("Accessing to db via ssl\n")
		s, err = dialUsingSSL(address, option, username, password)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			os.Exit(1)
		}
	}
	defer s.Close()

	// Create a new handler
	handler := gorest.NewHandler()

	// Define and setup your custom structures.
	var nnRes NNResource
	var statRes StatisticsResource
	go updateDatabaseInfosPeriodically(s, &nnRes, &statRes)

	var herosRes HeroResource

	// Register the routes.
	handler.SetRoutes([]*gorest.Route{
		gorest.NewRoute(&nnRes, "/nn/predict"),
		gorest.NewRoute(&herosRes, "/heros"),
		gorest.NewRoute(&statRes, "/stats"),
	})

	// Get the handler for your HTTP(S) server.
	router := handler.GetMuxRouter(nil)
	log.Printf("serving on 0.0.0.0:8080\n")
	http.ListenAndServe("0.0.0.0:8080", router)
}

func getActualNewNeuralNetwork(s *mgo.Session) (*rprop.NeuralNetwork, error) {
	var NN rprop.NeuralNetwork
	err := s.DB("neuralnetwork").C("weights").Find(nil).One(&NN)
	if err != nil {
		return nil, err
	}
	return &NN, nil
}

func getStatistics(s *mgo.Session) (*rprop.ValidationResult, error) {
	var stats rprop.ValidationResult
	err := s.DB("neuralnetwork").C("score").Find(nil).One(&stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func updateDatabaseInfosPeriodically(s *mgo.Session, nnRes *NNResource, statRes *StatisticsResource) {
	for {
		NN, err := getActualNewNeuralNetwork(s)
		if err != nil {
			log.Printf("%s\n", err.Error())
			os.Exit(1)
		}
		NN.ActivationFunction = rprop.Logistic
		NN.DerivateActivation = rprop.DerivateLogistic
		NN.ErrorFunction = rprop.SSE
		NN.DerivateError = rprop.DerivateSSE
		nnRes.nn = NN

		stats, err := getStatistics(s)
		if err != nil {
			log.Printf("%s\n", err.Error())
			os.Exit(1)
		}
		statRes.Stats = stats

		time.Sleep(5 * time.Minute)
	}
}

func dialUsingSSL(addresses string, dboption string, username string, password string) (*mgo.Session, error) {
	listaddresses := make([]string, 0)
	for _, str := range strings.Split(addresses, ",") {
		if str != "" {
			listaddresses = append(listaddresses, str)
		}
	}
	dboptions := strings.Split(dboption, "=")
	if len(dboption) < 2 {
		return nil, fmt.Errorf("can not found authSource keyword in order to permit SSL connection, aborting")
	}
	tlsConfig := &tls.Config{}
	dialInfo := &mgo.DialInfo{
		Addrs:    listaddresses,
		Database: dboptions[1],
		Username: username,
		Password: password,
	}

	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}
	session.EnsureSafe(&mgo.Safe{
		W:     1,
		FSync: false,
	})
	return session, nil
}
