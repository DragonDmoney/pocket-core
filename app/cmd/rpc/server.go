package rpc

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/pokt-network/pocket-core/app"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var APIVersion = app.AppVersion

func StartRPC(port string, timeout int64, simulation bool) {
	routes := GetRoutes()
	if simulation {
		simRoute := Route{Name: "SimulateRequest", Method: "POST", Path: "/v1/client/sim", HandlerFunc: SimRequest}
		routes = append(routes, simRoute)
	}

	srv := &http.Server{
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 20 * time.Second,
		WriteTimeout:      60 * time.Second,
		Addr:              ":" + port,
		Handler:           http.TimeoutHandler(Router(routes), time.Duration(timeout)*time.Millisecond, "Server Timeout Handling Request"),
	}
	log.Fatal(srv.ListenAndServe())
}

func Router(routes Routes) *httprouter.Router {
	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.Method, route.Path, route.HandlerFunc)
	}
	return router
}

func cors(w *http.ResponseWriter, r *http.Request) (isOptions bool) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	return ((*r).Method == "OPTIONS")
}

type Route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc httprouter.Handle
}

type Routes []Route

func GetRoutes() Routes {
	routes := Routes{
		Route{Name: "AppVersion", Method: "GET", Path: "/v1", HandlerFunc: Version},
		Route{Name: "HandleDispatch", Method: "POST", Path: "/v1/client/dispatch", HandlerFunc: Dispatch},
		Route{Name: "HandleDispatchCORS", Method: "OPTIONS", Path: "/v1/client/dispatch", HandlerFunc: Dispatch},
		Route{Name: "Service", Method: "POST", Path: "/v1/client/relay", HandlerFunc: Relay},
		Route{Name: "ServiceCORS", Method: "OPTIONS", Path: "/v1/client/relay", HandlerFunc: Relay},
		Route{Name: "Challenge", Method: "POST", Path: "/v1/client/challenge", HandlerFunc: Challenge},
		Route{Name: "ChallengeCORS", Method: "OPTIONS", Path: "/v1/client/challenge", HandlerFunc: Challenge},
		Route{Name: "SendRawTx", Method: "POST", Path: "/v1/client/rawtx", HandlerFunc: SendRawTx},
		Route{Name: "QueryBlock", Method: "POST", Path: "/v1/query/block", HandlerFunc: Block},
		Route{Name: "QueryTX", Method: "POST", Path: "/v1/query/tx", HandlerFunc: Tx},
		Route{Name: "QueryAccountTxs", Method: "POST", Path: "/v1/query/accounttxs", HandlerFunc: AccountTxs},
		Route{Name: "QueryBlockTxs", Method: "POST", Path: "/v1/query/blocktxs", HandlerFunc: BlockTxs},
		Route{Name: "QueryHeight", Method: "POST", Path: "/v1/query/height", HandlerFunc: Height},
		Route{Name: "QueryBalance", Method: "POST", Path: "/v1/query/balance", HandlerFunc: Balance},
		Route{Name: "QueryAccount", Method: "POST", Path: "/v1/query/account", HandlerFunc: Account},
		Route{Name: "QueryNodes", Method: "POST", Path: "/v1/query/nodes", HandlerFunc: Nodes},
		Route{Name: "QueryNode", Method: "POST", Path: "/v1/query/node", HandlerFunc: Node},
		Route{Name: "QueryNodeParams", Method: "POST", Path: "/v1/query/nodeparams", HandlerFunc: NodeParams},
		Route{Name: "QueryNodeClaims", Method: "POST", Path: "/v1/query/nodeclaims", HandlerFunc: NodeClaims},
		Route{Name: "QueryNodeClaim", Method: "POST", Path: "/v1/query/nodeclaim", HandlerFunc: NodeClaim},
		Route{Name: "QueryApps", Method: "POST", Path: "/v1/query/apps", HandlerFunc: Apps},
		Route{Name: "QueryApp", Method: "POST", Path: "/v1/query/app", HandlerFunc: App},
		Route{Name: "QueryAppParams", Method: "POST", Path: "/v1/query/appparams", HandlerFunc: AppParams},
		Route{Name: "QueryPocketParams", Method: "POST", Path: "/v1/query/pocketparams", HandlerFunc: PocketParams},
		Route{Name: "QuerySupportedChains", Method: "POST", Path: "/v1/query/supportedchains", HandlerFunc: SupportedChains},
		Route{Name: "QuerySupply", Method: "POST", Path: "/v1/query/supply", HandlerFunc: Supply},
		Route{Name: "QueryDAOOwner", Method: "POST", Path: "/v1/query/daoowner", HandlerFunc: DAOOwner},
		Route{Name: "QueryUpgrade", Method: "POST", Path: "/v1/query/upgrade", HandlerFunc: Upgrade},
		Route{Name: "QueryACL", Method: "POST", Path: "/v1/query/acl", HandlerFunc: ACL},
		Route{Name: "QueryAllParams", Method: "POST", Path: "/v1/query/allparams", HandlerFunc: AllParams},
		Route{Name: "QueryParam", Method: "POST", Path: "/v1/query/param", HandlerFunc: Param},
		Route{Name: "QueryState", Method: "POST", Path: "/v1/query/state", HandlerFunc: State},
	}
	return routes
}

func WriteResponse(w http.ResponseWriter, jsn, path, ip string) {
	b, err := json.Marshal(jsn)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		fmt.Println(err.Error())
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, err := w.Write(b)
		if err != nil {
			fmt.Println(fmt.Errorf("error in RPC Handler WriteResponse: %v", err))
		}
	}
}

func WriteRaw(w http.ResponseWriter, jsn, path, ip string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err := w.Write([]byte(jsn))
	if err != nil {
		fmt.Println(fmt.Errorf("error in RPC Handler WriteRaw: %v", err))
	}
}
func WriteJSONResponse(w http.ResponseWriter, jsn, path, ip string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsn), &raw); err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		fmt.Println(fmt.Errorf("error in RPC Handler WriteJSONResponse: %v", err))
		return
	}
	err := json.NewEncoder(w).Encode(raw)
	if err != nil {
		fmt.Println(fmt.Errorf("error in RPC Handler WriteJSONResponse: %v", err))
		return
	}
}

func WriteErrorResponse(w http.ResponseWriter, errorCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(errorCode)
	err := json.NewEncoder(w).Encode(&rpcError{
		Code:    errorCode,
		Message: errorMsg,
	})
	if err != nil {
		fmt.Println(fmt.Errorf("error in RPC Handler WriteErrorResponse: %v", err))
	}
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func PopModel(_ http.ResponseWriter, r *http.Request, _ httprouter.Params, model interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	if err := r.Body.Close(); err != nil {
		return err
	}
	if err := json.Unmarshal(body, model); err != nil {
		return err
	}
	return nil
}
