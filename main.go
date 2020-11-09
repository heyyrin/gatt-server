package main

import (
	"fmt"
	"github.com/paypal/gatt/examples/option"
	"github.com/paypal/gatt/examples/service"
	"log"
	"time"

	"github.com/paypal/gatt"
)

// sample service
func NewCountService() *gatt.Service {
	n := 0
	s := gatt.NewService(gatt.MustParseUUID("09fc95c0-c111-11e3-9904-0002a5d5c51b"))
	s.AddCharacteristic(gatt.MustParseUUID("11fac9e0-c111-11e3-9246-0002a5d5c51b")).HandleReadFunc(
		func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
			fmt.Fprintf(rsp, "count: %d", n)
			n++
		})

	s.AddCharacteristic(gatt.MustParseUUID("16fe0d80-c111-11e3-b8c8-0002a5d5c51b")).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			log.Println("Wrote:", string(data))
			return gatt.StatusSuccess
		})

	s.AddCharacteristic(gatt.MustParseUUID("1c927b50-c116-11e3-8a33-0800200c9a66")).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			cnt := 0
			for !n.Done() {
				fmt.Fprintf(n, "Count: %d", cnt)
				cnt++
				time.Sleep(time.Second)
			}
		})

	return s
}

var UART_SERVICE_UUID = "6E400001-B5A3-F393-E0A9-E50E24DCCA9E"
var UART_RX_CHAR_UUID = "6E400002-B5A3-F393-E0A9-E50E24DCCA9E" // RX Characteristic (Property = Notify)
var UART_TX_CHAR_UUID = "6E400003-B5A3-F393-E0A9-E50E24DCCA9E" // TX Characteristic (Property = Write without response)

func NewUartService() *gatt.Service {

	s := gatt.NewService(gatt.MustParseUUID(UART_SERVICE_UUID))
	s.AddCharacteristic(gatt.MustParseUUID(UART_RX_CHAR_UUID)).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			cnt := 0
			for !n.Done() { // notif
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

	s.AddCharacteristic(gatt.MustParseUUID(UART_TX_CHAR_UUID)).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			log.Println("Wrote:", string(data))
			return gatt.StatusSuccess
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
		gattServer.AdvertiseNameAndServices("HYERIN001", []gatt.UUID{uartService.UUID()}) //

		//gattServer.AdvertiseIBeacon(gatt.MustParseUUID("AA6062F098CA42118EC4193EB73CCEB6"), 1, 2, -59)  // beacon code!

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
			fmt.Println(c.MTU())
		}),
	)

	gattServer.Init(onStateChanged)

	select {}
}
