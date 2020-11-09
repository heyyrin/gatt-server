package main

import (
	"fmt"
	"github.com/paypal/gatt/examples/option"
	"github.com/paypal/gatt/examples/service"
	"log"
	"time"

	"github.com/paypal/gatt"
)

// example
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

// 6E400001-B5A3-F393-E0A9-E50E24DCCA9E for the Service
// 6E400002-B5A3-F393-E0A9-E50E24DCCA9E for the RX Characteristic (Property = Notify)
// 6E400003-B5A3-F393-E0A9-E50E24DCCA9E for the TX Characteristic (Property = Write without response)
func NewTestService() *gatt.Service {
	//n := 0
	s := gatt.NewService(gatt.MustParseUUID("6E400001-B5A3-F393-E0A9-E50E24DCCA9E"))

	//s.AddCharacteristic(gatt.MustParseUUID("6E400002-B5A3-F393-E0A9-E50E24DCCA9E")).HandleReadFunc(
	//	func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
	//		fmt.Fprintf(rsp, "count: %d", n)
	//		n++
	//	})

	s.AddCharacteristic(gatt.MustParseUUID("6E400002-B5A3-F393-E0A9-E50E24DCCA9E")).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) { // no
			fmt.Fprintf(n, "Count: %d", n.Cap())
		})

	s.AddCharacteristic(gatt.MustParseUUID("6E400003-B5A3-F393-E0A9-E50E24DCCA9E")).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			log.Println("Wrote:", string(data))
			return gatt.StatusSuccess
		})

	return s
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

	onStateChanged := func(d gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			// Setup GAP and GATT services
			gattServer.AddService(service.NewGapService("HYERIN001"))
			//gattServer.AddService(service.NewGattService())

			// add service
			testService := NewTestService()
			gattServer.AddService(testService)

			// Advertise device name and service's UUIDs.
			gattServer.AdvertiseNameAndServices("HYERIN001", []gatt.UUID{testService.UUID()}) // 해당 코드 있어야함

			//gattServer.AdvertiseIBeacon(gatt.MustParseUUID("AA6062F098CA42118EC4193EB73CCEB6"), 1, 2, -59)  // beacon code
		default:
			fmt.Print("default????")
		}
	}

	gattServer.Init(onStateChanged)

	select {}
}
