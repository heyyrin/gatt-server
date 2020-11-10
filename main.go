package main

import (
	"fmt"
	"github.com/paypal/gatt/examples/option"
	"github.com/paypal/gatt/examples/service"
	"log"
	"time"

	"github.com/paypal/gatt"
)

var UART_SERVICE_UUID = "6E400001-B5A3-F393-E0A9-E50E24DCCA9E"
var UART_RX_CHAR_UUID = "6E400002-B5A3-F393-E0A9-E50E24DCCA9E" // RX Characteristic (Property = Notify)
var UART_TX_CHAR_UUID = "6E400003-B5A3-F393-E0A9-E50E24DCCA9E" // TX Characteristic (Property = Write without response)

func NewUartService() *gatt.Service {

	s := gatt.NewService(gatt.MustParseUUID(UART_SERVICE_UUID))

	// Mobile -> Raspberry
	s.AddCharacteristic(gatt.MustParseUUID(UART_RX_CHAR_UUID)).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			log.Println("Wrote:", string(data))
			return gatt.StatusSuccess
		})

	// Raspberry -> Mobile
	s.AddCharacteristic(gatt.MustParseUUID(UART_TX_CHAR_UUID)).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			cnt := 0
			for !n.Done() {
				// 0과 1을 반복해서 데이터를 전송
				if cnt%2 == 0 {
					data := "1"
					sentLen, err := n.Write([]byte(data))
					if err != nil {
						log.Println(err)
					} else {
						log.Println("Send:", data, " Length :", sentLen)
					}
				} else {
					data := "0"
					sentLen, err := n.Write([]byte(data))
					if err != nil {
						log.Println(err)
					} else {
						log.Println("Send:", data, " Length :", sentLen)
					}
				}
				time.Sleep(time.Second)
				cnt++
			}
		})

	return s
}

func onStateChanged(gattServer gatt.Device, s gatt.State) {

	fmt.Printf("State: %s\n", s)
	switch s {
	case gatt.StatePoweredOn:
		// Setup GAP and GATT services
		gattServer.AddService(service.NewGapService("HYERIN001")) // 'Generic Access' Service
		gattServer.AddService(service.NewGattService())           // 'Generic Attribute' Service

		// add uart service
		uartService := NewUartService()
		gattServer.AddService(uartService)

		// Advertise device name and service's UUIDs.
		gattServer.AdvertiseNameAndServices("HYERIN001", []gatt.UUID{uartService.UUID()})

	default:
		fmt.Print("default????")
	}

}

func main() {

	gattServer, err := gatt.NewDevice(option.DefaultServerOptions...)
	if err != nil {
		panic(err)
	}

	// central conn/disconn handler
	gattServer.Handle(
		gatt.CentralConnected(func(c gatt.Central) { fmt.Println("Connect: ", c.ID()) }),
		gatt.CentralDisconnected(func(c gatt.Central) {
			fmt.Println("Disconnect: ", c.ID())
		}),
	)

	gattServer.Init(onStateChanged)

	select {}
}
