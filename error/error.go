package error

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

var ErrorMessages map[string]string

type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type ErrorReponse struct {
	Error *Error `json:"Error,omitempty"`
}

func (r *Error) Error() string {
	return fmt.Sprintf("code:%s;message:%s", r.Code, r.Message)
}

func DefaultErrorMessages() error {
	//Reading Json File
	jsonFile, err := os.Open("errorMessage.json")
	if err != nil {
		log.Error(err)
		return err
	}

	defer func() {
		cerr := jsonFile.Close()
		if err == nil {
			err = cerr
		}
	}()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &ErrorMessages)
	if err != nil {
		log.Error("File UnMarshall error ", err)
		return err
	}
	return nil
}
func HandleErrorResponse(err error) *ErrorReponse {
	resp := &ErrorReponse{Error: &Error{Code: "INTERNAL_ERROR", Message: "Internal Error"}}
	var errorMessage string
	errorMessage = err.Error()
	if len(ErrorMessages) == 0 || errorMessage == "" || !(strings.HasPrefix(errorMessage, "code:") || strings.HasPrefix(errorMessage, `{"Error`)) {

		return resp
	}

	if strings.HasPrefix(errorMessage, `{"Error`) {
		var res *ErrorReponse
		if err := json.Unmarshal([]byte(errorMessage), &res); err != nil {
			return resp
		}
		return res
	}
	messages := strings.Split(errorMessage, ";")
	if len(messages) < 2 {
		return resp
	}
	code := strings.ReplaceAll(messages[0], "code:", "")
	message := strings.ReplaceAll(messages[1], "message:", "")
	if code != "BAD_INPUT" && ErrorMessages[code] != "" {

		resp = &ErrorReponse{Error: &Error{Code: code, Message: ErrorMessages[code]}}
	}
	if code == "BAD_INPUT" {
		resp = &ErrorReponse{Error: &Error{Code: code, Message: message}}
	}
	return resp
}
