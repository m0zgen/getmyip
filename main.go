package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var enableLogging bool

const logFileName = "client_ips.log"

func logIP(ip string) {

	// If flag -log passed
	if !enableLogging {
		return
	}

	// Create or open log file
	file, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Log file open error:", err)
		return
	}
	defer file.Close()

	// Get current time with time format
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Record time and IP in to log
	logEntry := fmt.Sprintf("%s - %s\n", currentTime, ip)
	_, err = file.WriteString(logEntry)
	if err != nil {
		fmt.Println("Error writing to log file:", err)
	}
}

func getIP(r *http.Request) string {

	if cfConnectiongIP := r.Header.Get("Cf-Connecting-Ip"); cfConnectiongIP != "" {
		//fmt.Println("Cf-Connecting-Ip:")
		return cfConnectiongIP
	}

	// Check X-Forwarder header
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// If X-Forwarded-For exists, will use first IP-address from addresses list
		//fmt.Println("X-Forwarded-For:")
		return strings.Split(xForwardedFor, ",")[0]
		// Another method through parse:
		//fmt.Println(net.ParseIP(xForwardedFor))
	}

	// Check X-Real-IP header
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		//fmt.Println("X-Real-IP:")
		return xRealIP
	}

	// If header does not exist, use RemoteAddr
	ip := strings.Split(r.RemoteAddr, ":")[0]
	//fmt.Println("Remote addr:")
	return ip
}

func getServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "Can't determine server IP address"
	}

	// Используем первый доступный IP-адрес
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "Server IP does not found"
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Определяем IP-адрес клиента
	ip := getIP(r)

	// Выводим IP-адрес в консоль сервера
	fmt.Printf("Client IP: %s\n", ip)

	// Отправляем IP-адрес обратно клиенту
	fmt.Fprintf(w, "%s", ip)

	logIP(ip)
}

func main() {

	port := flag.Int("port", 8080, "Port for server")
	flag.BoolVar(&enableLogging, "log", false, "Enable IP logging")
	flag.Parse()

	// Определяем обработчик запросов
	http.HandleFunc("/", handleRequest)

	// Получаем и выводим IP-адрес сервера
	serverIP := getServerIP()
	fmt.Printf("Server IP: %s\n", serverIP)

	// Запускаем сервер на flag порту 8080
	addr := fmt.Sprintf(":%d", *port)

	fmt.Printf("Server run on http://%s%s\n", serverIP, addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Start server error:", err)
	}
}
