package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Id struct {
	Id string `json:"id"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Variable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type IncomingData struct {
	Id       string     `json:"id"`
	Variable []Variable `json:"variables"`
}

var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

func accessTokenCall() (accessToken interface{}, err error) {

	url := "https://34.93.102.191:18080/auth/realms/camunda-platform/protocol/openid-connect/token"
	method := "POST"

	payload := strings.NewReader("client_id=tasklist&client_secret=XALaRPl5qwTEItdwCMiPS62nVpKs7dL7&grant_type=client_credentials")
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("Error:%v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("Error:%v", err)
	}
	defer res.Body.Close()

	var Token_ResponseData map[string]interface{}
	json.NewDecoder(res.Body).Decode(&Token_ResponseData)

	return Token_ResponseData["access_token"].(string), nil
}

func getTasksHandler(response http.ResponseWriter, request *http.Request) {

	accessToken, err := accessTokenCall()
	// fmt.Println(accessToken)
	if err != nil || accessToken == nil {
		fmt.Println("Error In Getting Token:", err)
	}

	var idBody Id

	json.NewDecoder(request.Body).Decode(&idBody)
	if idBody.Id == "" {
		response.Header().Set("Content-Type", "application/json")
		json.NewEncoder(response).Encode(map[string]string{
			"Success": "False",
			"Message": "Do Provide Required Data",
		})
		fmt.Println("Do Provide Required Data")
		return
	}
	fmt.Println(idBody.Id)

	myid := idBody.Id

	url := "http://34.93.102.191:8082/v1/tasks/search"
	method := "POST"

	payload := fmt.Sprintf(`{
		"state": "CREATED",
		"assigned": true,
		"assignee": "%s"
	}`, myid)

	reader := strings.NewReader(payload)

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		fmt.Println(err)
		return
	}

	Token := fmt.Sprintf("Bearer %s", accessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", Token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	// data,err:=ioutil.ReadAll(res.Body)
	// // fmt.Println(err)
	// fmt.Println(string(data))

	// getData := string(data)

	var getTaskData []map[string]interface{}

	err = json.NewDecoder(res.Body).Decode(&getTaskData)

	if err != nil {
		fmt.Println("Error Decoding getTaskData:", err)
	}
	// fmt.Println(getTaskData)
	response.Header().Set("Content-Type", "application/json")

	if len(getTaskData) == 0 {
		json.NewEncoder(response).Encode(map[string]string{
			"Success": "True",
			"Message": "No Remaining Tasks",
		})
		return
	}
	json.NewEncoder(response).Encode(getTaskData)

}

func login(response http.ResponseWriter, request *http.Request) {

}

func ValidateLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var req LoginRequest
		err := json.NewDecoder(request.Body).Decode(&req)
		if err != nil {
			fmt.Println("Error While Decoding in Validation:", err)
			return
		}
		if req.Username == "" && req.Password == "" {
			next.ServeHTTP(response, request)
		} else {
			return
		}

	})
}
func fetchDataAndForm(response http.ResponseWriter, request *http.Request) {

	var requestBody map[string]interface{}
	json.NewDecoder(request.Body).Decode(&requestBody)
	if requestBody == nil {
		json.NewEncoder(response).Encode(map[string]string{
			"Success": "False",
			"Message": "Do Provide Required Data",
		})
		fmt.Println("Do Provide Required Data")
		return
	}

	id := requestBody["Id"].(string)
	formId := requestBody["formId"].(string)
	processDefinitionKey := requestBody["processDefinitionKey"].(string)
	formVersion := requestBody["formVersion"].(string)

	if id == "" || formId == "" || processDefinitionKey == "" || formVersion == "" {
		json.NewEncoder(response).Encode(map[string]string{
			"Success": "False",
			"Message": "Provide Required Details",
		})
		return
	}

	accessToken, err := accessTokenCall()
	if err != nil || accessToken == nil {
		fmt.Println("Error Getting Token:", err)
	}
	url := fmt.Sprintf("http://34.93.102.191:8082/v1/tasks/%s/variables/search", id)
	method := "POST"

	payload := strings.NewReader(``)

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	Token := fmt.Sprintf("Bearer %s", accessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", Token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	var taskVariable []map[string]string
	json.NewDecoder(res.Body).Decode(&taskVariable)
	// fmt.Println(taskVariable)

	// var extractedData map[string]string

	// }
	// fmt.Println(extractedData)

	// var extractedData map[string]interface{}
	extractedData := make(map[string]interface{})
	for _, value := range taskVariable {
		// extractedData[]
		key := value["name"]
		data := value["value"]
		extractedData[key] = data
	}
	fmt.Println(extractedData)

	//fetchFormSchema
	// var schemaData map[string]interface{}
	schemaData := fetchForm(accessToken.(string), formId, processDefinitionKey, formVersion)

	// obj:=schemaData[0]

	// components := (schemaData["components"]).([]interface{})

	// fmt.Println(components[0])

	// fmt.Println(schemaData)
	// json.NewDecoder(schema).Decode(schemaData)

	// fmt.Println(schema)
	response.Header().Set("Content-Type", "application/json")
	responseData := map[string]interface{}{
		"data":   extractedData,
		"schema": schemaData,
	}
	json.NewEncoder(response).Encode(responseData)

}
func fetchForm(acessToken string, formIDD string, processDefinitionKey string, formVersion string) map[string]interface{} {
	// FormVersion := fmt.Sprintf("%.0f", formVersion)
	url := fmt.Sprintf("http://34.93.102.191:8082/v1/forms/%s?processDefinitionKey=%s&version=%s", formIDD, processDefinitionKey, formVersion)
	// fmt.Println(url)

	method := "GET"

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	Token := fmt.Sprintf("Bearer %s", acessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", Token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error Making Request", err)
		return nil
	}
	defer res.Body.Close()

	var fetchdFormData map[string]string
	json.NewDecoder(res.Body).Decode(&fetchdFormData)

	var schemaData map[string]interface{}
	schema := fetchdFormData["schema"]
	json.Unmarshal([]byte(schema), &schemaData)
	return schemaData
}

func completeHandler(response http.ResponseWriter, request *http.Request) {

	Token, err := accessTokenCall()
	if err != nil {
		fmt.Println("Error Getting The Token")
	}
	defer request.Body.Close()
	var incData IncomingData
	err = json.NewDecoder(request.Body).Decode(&incData)
	if err != nil {
		fmt.Println("Error", err)
	}

	// fmt.Println(incData)

	// json.Unmarshal(data, &incData)
	if incData.Id == "" {
		json.NewEncoder(response).Encode(map[string]string{
			"Success": "False",
			"Message": "Invalid Id",
		})
		return
	}

	if err != nil || Token == nil {
		log.Println("Error:", err)
	}

	responseData := completeTask(&incData, Token.(string))
	response.Header().Set("Content-Type", "application/json")
	TaskState, ok := responseData["taskState"].(string)

	if TaskState == "" || !ok {
		errorMessage, _ := responseData["message"].(string)
		json.NewEncoder(response).Encode(map[string]string{
			"Success": "False",
			"Message": errorMessage,
		})
		log.Println("Unable To Complete The Task")
	} else {
		json.NewEncoder(response).Encode(responseData)
	}

}
func completeTask(incData *IncomingData, token string) map[string]interface{} {

	id := incData.Id
	url := fmt.Sprintf("http://34.93.102.191:8082/v1/tasks/%s/complete", id)
	method := "PATCH"

	x, err := json.Marshal(incData.Variable)
	if err != nil {
		fmt.Println("Error While Marshalling Data:", err)
	}

	D := "{\"variables\":" + string(x) + "}"
	reader := strings.NewReader(D)

	req, err := http.NewRequest(method, url, reader)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	accessToken := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", accessToken)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer res.Body.Close()

	var resposneDAta map[string]interface{}
	json.NewDecoder(res.Body).Decode(&resposneDAta)
	return resposneDAta

}

func testHandler(response http.ResponseWriter, request *http.Request) {
	var testData map[string]interface{}
	json.NewDecoder(request.Body).Decode(&testData)
	v := testData["data"].(map[string]interface{})

	var emptyMap []map[string]interface{}
	for key, value := range v {
		temp := make(map[string]interface{})
		temp["name"] = key
		temp["value"] = value
		emptyMap = append(emptyMap, temp)
	}

	fmt.Println(emptyMap)
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(emptyMap)
}
func main() {
	r := mux.NewRouter()
	s := r.PathPrefix("/api").Subrouter()
	s.Use(ValidateLogin)
	s.HandleFunc("/login", login).Methods("POST")
	r.HandleFunc("/getTasks", getTasksHandler).Methods("POST")
	r.HandleFunc("/fetchForm", fetchDataAndForm).Methods("POST")
	r.HandleFunc("/completeTask", completeHandler).Methods("POST")
	r.HandleFunc("/test", testHandler).Methods("POST")

	c := cors.Default()
	handler := c.Handler(r)

	port := ":4001"
	t := &http.Server{
		Addr:    port,
		Handler: handler,
	}
	log.Printf("Server is Running in Port%s", port)
	log.Fatal(t.ListenAndServe())
}

