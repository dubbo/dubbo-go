package extension

import "github.com/dubbo/dubbo-go/cluster"

var (
	loadbalances = make(map[string]func() cluster.LoadBalance)
)

func SetLoadbalance(name string, fcn func() cluster.LoadBalance) {
	loadbalances[name] = fcn
}

func GetLoadbalance(name string) cluster.LoadBalance {
	return loadbalances[name]()
}
