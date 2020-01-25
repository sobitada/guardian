package main

import (
    "errors"
    "flag"
    "fmt"
    "github.com/sirupsen/logrus"
    "github.com/sobitada/go-jormungandr/api"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "os"
    "time"
)

type NodeConfig struct {
    Rest struct {
        Listen string `yaml:"listen"`
    } `yaml:"rest"`
}

func getJormungandrAPI(nodeConfigPath string) (*api.JormungandrAPI, error) {
    var data, err = ioutil.ReadFile(nodeConfigPath)
    if err == nil {
        var nodeConfig NodeConfig
        err = yaml.Unmarshal(data, &nodeConfig)
        if err == nil {
            var host = nodeConfig.Rest.Listen
            if len(host) > 0 {
                return api.GetAPIFromHost(fmt.Sprintf("http://%v", host), 2*time.Second)
            } else {
                return nil, errors.New("host at which the API is listening must not be empty")
            }
        }
    }
    return nil, err
}

func main() {
    interval := flag.Int64("i", 1000, "interval in milliseconds.")
    flag.Parse()
    args := flag.Args()
    if len(args) == 1 && len(args[0]) > 0 {
        // logging
        logrus.SetLevel(logrus.InfoLevel)
        logrus.SetFormatter(&logrus.TextFormatter{
            FullTimestamp: true,
        })
        // read config and extract rest api url
        apiObj, err := getJormungandrAPI(args[0])
        if err == nil {
            // start guarding the node
            guard(apiObj, time.Duration(*interval)*time.Millisecond)
        } else {
            logrus.Errorf("API access could not be constructed given the node config at '%v'.", args[0])
            os.Exit(1)
        }
    } else {
        fmt.Printf("Usage: guardian <node-config.yml>\n")
        flag.PrintDefaults()
        os.Exit(1)
    }
}

func guard(api *api.JormungandrAPI, interval time.Duration) {
    for ; ; {
        stats, bootstrapping, err := api.GetNodeStatistics()
        if err == nil {
            if !bootstrapping && stats != nil {
                break
            } else if bootstrapping {
                logrus.Info("Node is bootstrapping.")
            }
        } else {
            logrus.Infof("Node is not reachable. %v", err.Error())
        }
        time.Sleep(interval)
    }
    // delete all registered leaders
    leaders, err := api.GetRegisteredLeaders()
    logrus.Infof("Registered leaders detected: %v", leaders)
    if err == nil {
        for n := range leaders {
            deleteLeader(api, leaders[n])
        }
    } else {
        logrus.Error("Failed to deregister the leaders.")
    }
}

func deleteLeader(api *api.JormungandrAPI, leaderID uint64) {
    for i := 0; i < 3; i++ {
        logrus.Infof("%v. try to remove the leader %v", i+1, leaderID)
        found, err := api.RemoveRegisteredLeader(leaderID)
        if found {
            logrus.Infof("Deleted leader %v successfully.", leaderID)
            break
        } else if err != nil {
            logrus.Errorf("Leader %v: %v", leaderID, err.Error())
        }
        time.Sleep(1 * time.Second)
    }
}
