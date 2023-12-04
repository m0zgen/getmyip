package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

func getIP(r *http.Request) string {
	// Check X-Real-IP header
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}
	// Check X-Forwarder header
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// Если заголовок X-Forwarded-For присутствует, используем первый IP-адрес в списке
		return strings.Split(xForwardedFor, ",")[0]
		// Another method through parse:
		//fmt.Println(net.ParseIP(xForwardedFor))
	}

	if cfConnectiongIP := r.Header.Get("Cf-Connecting-Ip"); cfConnectiongIP != "" {
		return cfConnectiongIP
	}

	// Если заголовок отсутствует, используем RemoteAddr
	ip := strings.Split(r.RemoteAddr, ":")[0]
	return ip
}

func getServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "Не удалось определить IP сервера"
	}

	// Используем первый доступный IP-адрес
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "IP сервера не найден"
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Определяем IP-адрес клиента
	ip := getIP(r)

	// Выводим IP-адрес в консоль сервера
	fmt.Printf("IP клиента: %s\n", ip)

	// Отправляем IP-адрес обратно клиенту
	fmt.Fprintf(w, "%s", ip)
}

func main() {
	// Определяем обработчик запросов
	http.HandleFunc("/", handleRequest)

	// Получаем и выводим IP-адрес сервера
	serverIP := getServerIP()
	fmt.Printf("IP сервера: %s\n", serverIP)

	// Запускаем сервер на порту 8080
	port := 8080
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("Сервер запущен на http://%s%s\n", serverIP, addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
	}
}
