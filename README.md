# udpmon

*udpmon* allows to monitor the performance and latency of network connections.
It consists of a client and a server. The server just returns every udp package
it receives back to the sender. 

The client can be configured to send udp packages to the server with a configurable
frequency and wait for the server to send the package back. If the package did not
come back after the timeout, an error is logged. As soon as the packages start returning
again, it is logged as well. 

It is possible to configure a latency thershold and a latency logfile. In this latency
logifle any package that took longer than the latency thershold but shorter than the timeout
is logged.

## Useage

    udpmon OPTIONS

*Options:*

    -frequency int
        specify the amount of time between requests in millisecons (default 100)
    -latencylogfile string
        specify a file to store latency logs
    -latencythreshold int
        specify the latency threshold starting from which it is logged in the latency logfile (default 80)
    -listen string
        specify an address to listen on
    -logfile string
        specify a file to write the logs to
    -server string
        specify the destination server
    -timeout int
        specify the timeout for round trips in milliseconds (defaul 100)
