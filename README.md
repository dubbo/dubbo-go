# dubbo-go #
---
Apache Dubbo Golang Implementation.

## License

Apache License, Version 2.0

## Code design ##
Based on dubbo's layered code design (protocol layer,registry layer,cluster layer,config layer and so on),

About detail design please refer to [code layered design](https://github.com/dubbo/go-for-apache-dubbo/wiki/dubbo-go-V2.6-design)
## Feature list ##

+  Role: Consumer(√), Provider(√)

+  Transport: HTTP(√), TCP(√) Based on [getty](https://github.com/AlexStocks/getty)

+  Codec:  JsonRPC(√), Hessian(√) Based on [Hession2](https://github.com/dubbogo/hessian2)

+  Registry: ZooKeeper(√)

+  Cluster Strategy: Failover(√)

+  Load Balance: Random(√)

+  Filter: Echo(√)

## Code Example

The subdirectory examples shows how to use dubbo-go. Please read the examples/readme.md carefully to learn how to dispose the configuration and compile the program.


## Todo list

Implement more extention:

 * cluster strategy : Failfast/Failsafe/Failback/Forking/Broadcast

 * load balance strategy: RoundRobin/LeastActive/ConsistentHash

 * standard filter in dubbo: TokenFilter/AccessLogFilter/CountFilter/ActiveLimitFilter/ExecuteLimitFilter/GenericFilter/TpsLimitFilter

 * registry impl: consul/etcd/k8s
 
Compatible with dubbo v2.7.x and not finished function in dubbo v2.6.x:
 
 * routing rule (dubbo v2.6.x)
 
 * monitoring (dubbo v2.6.x)
 
 * metrics (dubbo v2.6.x)
 
 * dynamic configuration (dubbo v2.7.x)
