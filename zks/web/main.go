package web

import (
	"github.com/samuel/go-zookeeper/zk"

	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

var (
	argZk  = flag.String("zks", "127.0.0.1", "zookeeper servr URL")
	argWeb = flag.String("web", ":8080", "web server URI ([{host}]:{port})")

	conn *zk.Conn
)

type Node struct {
	Children []string
	Stat     *zk.Stat
}

func hRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8"/>
    <title>ZooKeeper UI</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/angular.js/1.5.8/angular.min.js"></script>
    <script>
    var app = angular.module("zkUi", []);
    app.controller("main", function($scope, $http){
        var path = $scope.path = [];
        var refresh = $scope.refresh = function(n) {
            $http.get("/api/" + path.join("/"))
                .then(function(res){
                    $scope.node = res.data;
                    })
        };
        refresh();
        $scope.select = function(child) {
            path.push(child);
            refresh();
        }
        $scope.back = function(n) {
            if( angular.isUndefined(n)){
                path.length = 0;
            } else {
                for(var i = path.length; i > n; i--) {
                    path.pop();
                }
            }
            refresh();
        }
    });
    </script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/skeleton/2.0.4/skeleton.min.css"/>
</head>
<body ng-app="zkUi" ng-controller="main">
    <div class="container">
    <h1>ZooKeeper UI</h1>
    <p>Manage here your ZooKeeper instance.</p>
    <p>
        <table>
            <tr>
                <th>Path</th>
                <td>
                    <span>
                        <a href="" ng-click="back()">(root)</a>
                    </span>
                    <span ng-repeat="p in path">
                        / <a href="" ng-click="back($index + 1)">{{p}}</a>
                    </span>
                </td>
            </tr>
            <tr>
                <th>Ctime</th>
                <td>{{node.Stat.Ctime}}</td>
            </tr>
            <tr>
                <th>CVersion</th>
                <td>{{node.Stat.Cversion}}</td>
            </tr>
            <tr>
                <th>A Version</th>
                <td>{{node.Stat.Aversion}}</td>
            </tr>
            <tr>
                <th>Version</th>
                <td>{{node.Stat.Version}}</td>
            </tr>
            <tr>
                <th># Children</th>
                <td>{{node.Stat.NumChildren}}</td>
            </tr>
        </table>{{node}}

    </p>
    <p>
        <button ng-click="back()">back</button>
        <ul>
            <li ng-repeat="c in node.Children">
                <a href="" ng-click="select(c)">{{c}}</a>
            </li>
        </ul>
    </p>
    </div>
</body>
</html>
`)
}
func hApi(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	children, stat, err := conn.Children(path)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error: %v", err)
		return
	}

	json.NewEncoder(w).Encode(Node{children, stat})
}

func main() {
	var err error
	log.Println("starting")

	defer log.Println("stopping")

	flag.Parse()

	// *** ZK ***
	log.Printf("connecting to ZooKeeper on %s...", *argZk)
	conn, _, err = zk.Connect([]string{*argZk}, time.Second) //*10)
	if err != nil {
		panic(err)
	}
	log.Printf("...done")

	// *** Web ***
	log.Printf("starting web UI on %s...", *argWeb)

	r := mux.NewRouter()
	r.HandleFunc("/api{path:.*}", hApi)
	r.HandleFunc("/", hRoot)
	http.Handle("/", r)
	err = http.ListenAndServe(*argWeb, nil)
	if err != nil {
		panic(err)
	}
	log.Printf("...done")

}

func Run(){
	main()
}
