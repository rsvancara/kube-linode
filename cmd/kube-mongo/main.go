package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/rs/zerolog/log"

	"os/signal"

	"github.com/coreos/go-iptables/iptables"
)

func BuildMongoChain(ipList []net.IP) {

	log.Info().Msg("building mongodb chain")
	ipt, err := iptables.New()
	if err != nil {
		log.Error().Err(err)
	}

	// Check if we have the chain
	ok, err := ipt.ChainExists("filter", "mongodb")
	if err != nil {
		log.Error().Err(err)
	}

	// clear the chain if exists, else create a new chain
	if ok {

		err = ipt.ClearChain("filter", "mongodb")
		if err != nil {
			log.Error().Err(err)
		}

	} else {

		err = ipt.NewChain("filter", "mongodb")
		if err != nil {
			log.Error().Err(err)
		}
	}

	for _, i := range ipList {
		//-s 1.2.3.4 -p tcp -m tcp --dport 27017
		err = ipt.Append("filter", "mongodb", "-s", i.String(), "-p", "tcp", "-m", "tcp", "--dport", "27017", "-j", "ACCEPT")
		if err != nil {
			log.Error().Err(err)
		}
	}

	rules, err := ipt.List("filter", "mongodb")
	if err != nil {
		log.Error().Err(err)
	}

	for _, v := range rules {
		log.Info().Msgf("configure rule: %s", v)
	}
}

func getKubeNodes(kubeconfig *string) ([]net.IP, error) {

	log.Info().Msg("querying kubernetes for node list")

	var results []net.IP

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Error().Err(err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error().Err(err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error().Err(err)
	}
	available := 0
	for _, val := range nodes.Items {
		//	fmt.Print("-----\n\n")

		if strIP, ok := val.Annotations["projectcalico.org/IPv4Address"]; ok {

			IPAddress := net.ParseIP(strings.Split(strIP, "/")[0])
			log.Info().Msgf("found node: %s", IPAddress.String()) //do something here
			results = append(results, IPAddress)
			available = available + 1
		}
	}
	log.Info().Msgf("There are %d nodes in the cluster, of which %d are available", len(nodes.Items), available)

	return results, nil
}

func isDiff(oldHosts []net.IP, newHosts []net.IP) bool {

	log.Info().Msg("checking if differences exist from last node query")
	// Check to see if the host list has changed from last time.
	// Easy check is to look for size differences in array length
	if len(newHosts) != len(oldHosts) {
		log.Info().Msgf("node count changed from %d to %d", len(newHosts), len(oldHosts))
		return true
	}

	// Harder check, see if the if the list contains different addresses
	// by checking if we can find the address in one list in another list
	matches := 0
	for _, v := range oldHosts {
		for _, k := range newHosts {
			if v.String() == k.String() {
				matches = matches + 1
				break
			}
		}
	}

	// Matches must equal the number of array elements, means that we found all the matches
	if matches != len(newHosts) {
		log.Info().Msgf("lists do not match, found  %d matches for  %d records", matches, len(oldHosts))
		return true
	}

	log.Info().Msg("no changes detected in kubernetes nodes")

	return false
}

func main() {

	log.Info().Msg("Starting ")

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	go func() {
		// Track changes in the list
		var oldHosts []net.IP
		var newHosts []net.IP

		for {

			newHosts, _ = getKubeNodes(kubeconfig)
			if isDiff(newHosts, oldHosts) {

				BuildMongoChain(newHosts)

				time.Sleep(5 * time.Second)
			}

			// Reset for the next iteration
			oldHosts = newHosts

			time.Sleep(5 * time.Second)
		}
	}()

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c

	// The signal is received, you can now do the cleanup
	fmt.Println("Got signal:", s)

}
