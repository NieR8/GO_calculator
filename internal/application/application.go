package application

import (
	"encoding/json"
	"fmt"
	"github.com/NieR8/myProject/pkg/calculation"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

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

func jsnResult(w http.ResponseWriter, str string) { // формирует результат запроса в виде json и выводит его
	res := ResponseErr{ResErr: str}
	byts, _ := json.Marshal(res)
	w.Write(byts)
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {

	log := logrus.New()
	if r.Method != "POST" {
		w.WriteHeader(http.StatusInternalServerError) // 500 ошибка
		jsnResult(w, ErrInternalServer.Error())
		log.Errorf(ErrInternalServer.Error())
		return
	}
	request := &Request{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		jsnResult(w, err.Error())
		log.Errorf(err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if request.Expression == "" {
		w.WriteHeader(http.StatusInternalServerError) // 500 ошибка
		jsnResult(w, ErrInvalidExprName.Error())
		log.Errorf(ErrInvalidExprName.Error())
		return
	}
	result, err1 := calculation.Calc(request.Expression)
	if err1 != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		jsnResult(w, err1.Error())
		log.Errorf("ошибка: %s", err1.Error())
	} else {
		w.WriteHeader(http.StatusOK)
		jsnResult(w, fmt.Sprintf("%.3f", result))
		log.Infof("ответ: %.3f", result)
	}
}

func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}
