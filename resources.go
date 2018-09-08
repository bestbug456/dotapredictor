package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	rprop "github.com/bestbug456/gorpropplus"
	gorest "github.com/fredmaggiowski/gorest"
)

type NNResource struct {
	nn *rprop.NeuralNetwork
}

type PredictionRequest struct {
	Heros []string
}

type PredictionResponse struct {
	Win  int
	Prob float64
}

func (p *NNResource) Post(r *http.Request) (int, gorest.Response) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Can't read the body: %s.\n", err.Error())
		return http.StatusInternalServerError, nil
	}

	var request PredictionRequest

	err = json.Unmarshal(b, &request)
	if err != nil {
		log.Printf("Can't marshal heros: %s.\n", err.Error())
		return http.StatusInternalServerError, nil
	}

	picks := make([]int, 10)
	for i := 0; i < len(request.Heros); i++ {
		picks[i] = heroID[request.Heros[i]]
	}
	input := orderPickByTeamAndCreateBitmask(picks)

	ris, err := p.nn.Predict(input)
	if err != nil {
		log.Printf("Can't marshal heros: %s.\n", err.Error())
		return http.StatusInternalServerError, nil
	}

	var prob float64
	var win int
	if int(ris[0]) == 0 {
		prob = 1 - ris[0]
	} else {
		win = 1
		prob = ris[0]
	}

	out, err := json.Marshal(PredictionResponse{
		Prob: prob,
		Win:  win,
	})
	if err != nil {
		log.Printf("Can't marshal heros: %s.\n", err.Error())
		return http.StatusInternalServerError, nil
	}
	response := gorest.NewStandardResponse()
	response.SetBody(out)

	return http.StatusOK, response
}

func orderPickByTeamAndCreateBitmask(picks []int) []float64 {
	team1Pick := []int{
		picks[0],
		picks[3],
		picks[5],
		picks[7],
		picks[8],
	}
	team2Pick := []int{
		picks[1],
		picks[2],
		picks[4],
		picks[6],
		picks[9],
	}
	bitmasks := createBitmasksForTeam(team1Pick)
	supp := createBitmasksForTeam(team2Pick)
	bitmasks = append(bitmasks, supp...)
	return bitmasks
}

func createBitmasksForTeam(team []int) []float64 {
	bitmasks := make([]float64, 115)
	for i := 0; i < len(team); i++ {
		bitmasks[team[i]] = 1
	}
	return bitmasks
}

type HeroResource struct{}

func (p *HeroResource) Get(r *http.Request) (int, gorest.Response) {

	// Return the array of avaiable dota heros
	out, err := json.Marshal(heros)
	if err != nil {
		log.Printf("Can't marshal heros: %s.\n", err.Error())
		return http.StatusInternalServerError, nil
	}
	response := gorest.NewStandardResponse()
	response.SetBody(out)
	return http.StatusOK, response
}

type StatisticsResource struct {
	Stats *NNStats
}

type stats struct {
	Accuracy float64
}

func (p *StatisticsResource) Get(r *http.Request) (int, gorest.Response) {

	var statistics stats

	statistics.Accuracy = float64(p.Stats.MatrixQA[0]) / float64(p.Stats.MatrixQA[1])

	out, err := json.Marshal(statistics)
	if err != nil {
		log.Printf("Can't marshal stats: %s.\n", err.Error())
		return http.StatusInternalServerError, nil
	}
	response := gorest.NewStandardResponse()
	response.SetBody(out)
	return http.StatusOK, response
}
