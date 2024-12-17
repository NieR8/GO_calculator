package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NieR8/myProject/pkg/calculation"
	"net/http"
	"os"
)

var ErrInvalidExprName = errors.New("в теле запроса ключ выражения должен называться expression")

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	config := &Config{}
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

type Application struct {
	config *Config
}

func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result string `json:"result"`
}
type ResponseErr struct {
	ResErr string `json:"error"`
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {

	request := &Request{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		fmt.Printf(err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if request.Expression == "" {
		w.WriteHeader(http.StatusInternalServerError) // 500 ошибка
		resErr := ResponseErr{ResErr: ErrInvalidExprName.Error()}
		btsErr, _ := json.Marshal(resErr)
		w.Write(btsErr)
		//http.Error(w, ErrInvalidExprName.Error(), http.StatusUnprocessableEntity)
		fmt.Printf(ErrInvalidExprName.Error() + "\n")
		return
	}
	result, err := calculation.Calc(request.Expression)
	resJSN := Response{Result: fmt.Sprintf("%.3f", result)}
	byts, _ := json.Marshal(resJSN)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		resErr := ResponseErr{ResErr: err.Error()}
		btsErr, _ := json.Marshal(resErr)
		w.Write(btsErr)
		//fmt.Fprintf(w, "ошибка: %s", err.Error())
		fmt.Printf("ошибка: %s \n", err.Error())
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(byts)
		//fmt.Fprintf(w, "ответ: %.3f", result)
		fmt.Printf("ответ: %.3f \n", result)
	}
}

func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}
