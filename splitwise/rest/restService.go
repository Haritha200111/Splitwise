package rest

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	Error "split/error"
	"split/splitwise/dto"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc/status"
)

type ECDSASignature struct {
	R, S *big.Int
}

var router *mux.Router
var services map[string]interface{} = make(map[string]interface{})

func Init() {
	log.Println("Adaptor Init called ...")
	listen()
}

func RegisterService(serviceName string, service interface{}) {
	services[serviceName] = service
}
func init() {
	router = mux.NewRouter()
}

func listen() {
	corsMw := mux.CORSMethodMiddleware(router)
	router.Use(corsMw)

	// Register the login route (e.g., /login)
	router.Handle("/login", http.HandlerFunc(LoginHandler)).Methods("POST")

	// Other routes for handling services and methods
	contextPath := viper.GetString("ContextPath")
	// router.Handle(contextPath+"/{service}/{method}", http.HandlerFunc(handle))
	router.Handle(contextPath+"/SplitWiseService/AddUser", http.HandlerFunc(handle)).Methods("POST")
	router.Handle(contextPath+"/SplitWiseService/AddUserToGroup", http.HandlerFunc(handle)).Methods("POST")
	router.Handle(contextPath+"/SplitWiseService/CreateGroup", http.HandlerFunc(handle)).Methods("POST")
	router.Handle(contextPath+"/SplitWiseService/DeleteGroup", http.HandlerFunc(handle)).Methods("DELETE")
	router.Handle(contextPath+"/SplitWiseService/Payment", http.HandlerFunc(handle)).Methods("PUT")
	router.Handle(contextPath+"/SplitWiseService/Split", http.HandlerFunc(handle)).Methods("POST")

	// Set CORS headers and options
	headersOK := handlers.AllowedHeaders([]string{
		"Content-Type", "access-control-allow-origin", "access-control-request-headers",
		"Access-Control-Allow-Headers", "Access-Control-Allow-Methods",
		"Authorization", "X-Requested-With", "X-CSRF-Token",
	})
	originsOK := handlers.AllowedOrigins([]string{"*"})
	methodsOK := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT"})

	// Wrap the router with CORS settings
	loggedRouter := handlers.CORS(headersOK, originsOK, methodsOK)(router)
	listenAddr := ":5050"
	log.Println("server is listening on localhost:5040")

	// Start the server
	err := http.ListenAndServe(listenAddr, loggedRouter)
	if err != nil {
		log.Fatal("Http Error :", err)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the login request (e.g., JSON body with email and password)
	var loginRequest dto.UserAccount
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Println("loginRequestloginRequestloginRequestloginRequestloginRequestloginRequestloginRequestloginRequestloginRequest", loginRequest.EmailId)

	_, err := validateUser(loginRequest.EmailId)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(loginRequest.EmailId)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Respond with the token
	response := map[string]string{"token": token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GenerateJWT(email string) (string, error) {
	secretKey := viper.GetString("JWTSignedString")

	claims := jwt.MapClaims{
		"userEmail": email,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // Token expiration time (e.g., 24 hours)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func handle(w http.ResponseWriter, r *http.Request) {
	log.Println("rea****************", r.URL.Path)
	a := strings.Split(r.URL.String(), "/")
	log.Println("aaaaaaaaaa", a[1])
	var resp interface{}
	isJson := false

	rr := make(map[string]interface{})
	rr["MetaData"] = r.Header
	rr["RequestTime"] = time.Now()
	rr["ipAddress"] = getIPAddress(r)

	body, _ := ioutil.ReadAll(r.Body)
	rr["RequestBody"] = body
	requestName := a[2] + ":" + a[3]
	rr["ServiceName"] = a[2]
	rr["MethodName"] = a[3]
	rr["RequestName"] = requestName
	rr["isRpc"] = true
	var ctx context.Context
	resp, ctx = handleRpc(rr)
	if ctx != nil {
		pvtKey, _ := ctx.Value("ResponsePvtKey").(string)
		keyID, _ := ctx.Value("Key").(string)
		if pvtKey != "" && keyID != "" {
			if signature, err := Sign(resp, pvtKey, keyID); err == nil {
				w.Header().Set("Signature", signature)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if isJson {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("Json encode error: ", err)
		}
	} else {
		res, _ := resp.([]byte)
		_, _ = w.Write(res)
	}
}

func handleRpc(rr map[string]interface{}) (resp interface{}, ctx context.Context) {
	var err error
	var c interface{}
	if c, err = loggingMiddleware(rr); err != nil {
		log.Error("Handle middelware error: ", err)
	}
	if c != nil {
		ctx, _ = c.(context.Context)
	}

	mfp := make(map[string]interface{})
	if ctx != nil {
		ctx = context.WithValue(ctx, "MFP", mfp)
	}
	log.Println("ctx***********", ctx)
	if err == nil {
		log.Infof("ServiceName: %s, MethodName: %s, StartTime: %s, In %d millisecond", rr["ServiceName"], rr["MethodName"], time.Now().String(), 0)
		body, _ := rr["RequestBody"].([]byte)
		services, _ := rr["ServiceName"].(string)
		method, _ := rr["MethodName"].(string)
		if resp, err = invoke(services, method, string(body), ctx); err != nil {
			log.Error("Handle invoke error: ", err)
		}
		requestTime, _ := rr["RequestTime"].(time.Time)
		log.Infof("ServiceName: %s, MethodName: %s, EndTime: %s, In %f millisecond", rr["ServiceName"], rr["MethodName"], time.Now().String(), (time.Now().Sub(requestTime).Seconds() * 1000))
	}

	if err != nil {
		resp = Error.HandleErrorResponse(err)
	}
	return resp, ctx
}

func invoke(serviceName, methodName, payload string, ctx context.Context) (interface{}, error) {
	if services[serviceName] == nil {
		log.Error("Service not found: ", serviceName)
		return nil, Error.INTERNAL_ERROR
	}

	sName := reflect.New(reflect.TypeOf(services[serviceName]).Elem()).Interface()
	method := reflect.ValueOf(sName).MethodByName(methodName)

	if !method.IsValid() {
		log.Error("Method not found: ", serviceName+":"+methodName)
		return nil, Error.INTERNAL_ERROR
	}

	mtype := method.Type()
	args := mtype.NumIn()
	var tp reflect.Type

	if args == 2 {
		tp = mtype.In(1) // Payload type
	} else {
		tp = mtype.In(0) // Payload type
	}

	par := reflect.New(tp).Interface()
	err := json.Unmarshal([]byte(payload), par)
	if err != nil {
		log.Errorf("Error unmarshalling payload: %v", err)
		return nil, Error.InvalidParameter("Invalid Request")
	}

	v := reflect.ValueOf(par)

	if v.IsZero() {
		log.Error("Parameter is zero value after unmarshalling")
		return nil, Error.InvalidParameter("Invalid Request")
	}

	var vals []reflect.Value
	if args == 2 {
		h := reflect.ValueOf(ctx)
		log.Println("h", h, v.Elem())
		vals = method.Call([]reflect.Value{h, v.Elem()})
	} else {
		log.Println("h", v.Elem())
		vals = method.Call([]reflect.Value{v.Elem()})
	}

	log.Infof("Method called successfully, Response: %v", vals)

	if len(vals) > 1 {
		respErr, ok := vals[1].Interface().(error)
		if ok && respErr != nil {
			return nil, respErr
		}
	}

	return vals[0].Interface(), nil
}

func loggingMiddleware(req map[string]interface{}) (context.Context, error) {
	log.Println("%%%%%%%%%%%%%%%%%%%%%%", req["MetaData"])
	request := &dto.RPCMessage{}
	data, _ := json.Marshal(req)
	err := json.Unmarshal(data, &request)
	log.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^", len(request.MetaData["Authorization"]))
	if err != nil {
		log.Info("error on loggingMiddleware unMarshal ..")
		return nil, err
	}
	log.Println("request.RequestNamerequest.RequestName", request.RequestName)
	ipAddress := request.IpAddress
	if request.RequestName != "UserService:GetToken" && request.RequestName != "UserService:RegisterDeviceRequest" &&
		request.RequestName != "UserService:RegisterDevice" && request.RequestName != "UserService:ValidateOTP" && request.RequestName != "KeyService:InitiateRegisterDevice" &&
		request.RequestName != "SplitWiseService:AddUser" && request.RequestName != "KeyService:AddApplicationKey" && request.RequestName != "UserService:RegisterNewDevice" && request.RequestName != "UserService:RequestRegisterDevice" {
		log.Println("Service Authenticate Called")

		digest := sha256.Sum256(request.RequestBody)
		authnReq := dto.AuthenticateRequest{
			ServiceName:   request.ServiceName,
			MethodName:    request.MethodName,
			RequestName:   request.RequestName,
			RequestDigest: digest[:],
		}
		if v := request.MetaData["Authorization"]; len(v) > 0 {
			authnReq.Authorization = v[0]
		}
		if v := request.MetaData["Signature"]; len(v) > 0 {
			authnReq.RequestSignature = v[0]
		}
		if v := request.MetaData["User-Agent"]; len(v) > 0 {
			authnReq.UserAgent = v[0]
		}
		log.Println("authnReq.Authorization.................", authnReq.Authorization)
		log.Debug("signature: ", authnReq.RequestSignature)
		var err error
		resp, err := Authenticate(&authnReq)
		if err != nil {
			if s, ok := status.FromError(err); ok {
				return nil, errors.New(s.Message())
			}
			return nil, err
		}
		log.Println("auth response", resp)
		// user := resp.GetUser()
		// username := user.Name
		// if username == "" {
		// 	username = user.FirstName
		// 	if user.LastName != "" {
		// 		username = username + " " + user.LastName
		// 	}
		// }
		// user.UserName = username

		ctx := context.WithValue(context.Background(), "User", resp)
		if authnReq.RequestSignature != "" {
			ctx = context.WithValue(ctx, "signature", authnReq.RequestSignature)
		}
		ctx = context.WithValue(ctx, "useremail", resp.EmailId)
		// var roles []string
		// for _, role := range user.Roles {
		// 	roles = append(roles, role.Role)
		// }
		// ctx = context.WithValue(ctx, "roles", roles)
		// ctx = context.WithValue(ctx, "UserRoles", strings.Join(roles, ","))
		ctx = context.WithValue(ctx, "Headers", request.MetaData)
		ctx = context.WithValue(ctx, "Application", "PL")
		ctx = context.WithValue(ctx, "IP", ipAddress)
		ctx = context.WithValue(ctx, "requestBody", request.RequestBody)
		ctx = context.WithValue(ctx, "Authorization", authnReq.Authorization)
		ctx = context.WithValue(ctx, "Signature", authnReq.RequestSignature)

		return ctx, nil
	}
	return nil, nil
}

func getIPAddress(r *http.Request) string {
	ipAddress, _, _ := net.SplitHostPort(r.RemoteAddr)
	if !(r.Header.Get("X-Forwarded-For") == "") {
		ips := strings.Split(r.Header.Get("X-Forwarded-For"), ", ")
		ipAddress = ips[0]
	}
	return ipAddress
}

func Sign(payload interface{}, pvtKey string, key string) (string, error) {
	byt, err := json.Marshal(payload)
	if err != nil {
		log.Println("json Marshal error")
	}
	pk, err := DecodePrivateKeyFromString(pvtKey)
	if err != nil {
		log.Printf("Error while decode provate key: %v", err)
		return "", err
	}

	hash := sha256.Sum256(byt)
	r, s, err := ecdsa.Sign(rand.Reader, pk, hash[:])
	if err != nil {
		log.Println(err)
		return "", err
	}
	b, err := asn1.Marshal(ECDSASignature{r, s})
	if err != nil {
		log.Println(err)
		return "", err
	}
	signature := "keyId=" + key + ",algorithm=ecdsa-sha256," + base64.StdEncoding.EncodeToString(b)
	return signature, nil
}
func DecodePrivateKeyFromString(pvtKey string) (*ecdsa.PrivateKey, error) {

	block, _ := pem.Decode([]byte(pvtKey))
	if block == nil {
		return nil, errors.New("failed to decode PEM private key")
	}
	x509Encoded := block.Bytes
	privateKey, err := x509.ParsePKCS8PrivateKey(x509Encoded)
	if err != nil {
		return nil, errors.New("failed to parse ECDSA private key")
	}
	switch privateKey := privateKey.(type) {
	case *ecdsa.PrivateKey:
		return privateKey, nil
	}
	return nil, errors.New("unsupported public key type")
}
