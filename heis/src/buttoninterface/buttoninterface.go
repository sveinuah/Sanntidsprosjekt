package elevdriver

//fix imports
import . "./src/elev"

type OrderType struct {
	Floor     int
	Direction int
	Argument  bool
}

var buttonList [][]bool

//Should this module be able to receive buttonList type arguments?
func buttonInterface(downChan chan OrderType, upChan chan OrderType, numFloors int) {
	buttonInterfaceInit(numFloors)
	for {
		//Get new button presses and send order up
		for floor := 0; floor < numFloors; i++ {
			for dir := 0; dir < 3; j++ {
				if elevGetButtonSignal(floor, dir) == true && lightList[floor][dir] == false { //Make variable lightList or read hardware each time?
					var order OrderType
					order.Floor = floor
					order.Direction = dir
					upChan <- order
				}
			}
		}
		//Get new orders and set/clear lights
		ordersInChannel := true
		for ordersInChannel {
			select {
			case order := <-downChan:
				elevButtonLight(order.Floor, order.Direction, order.Argument)
				buttonList[Order.Floor][order.Direction] = order.Argument
			default:
				ordersInChannel = false
			}
		}
	}
}

func buttonInterfaceInit(numFloors int) {
	for floor := 0; floor < numFloors; i++ {
		for dir := 0; dir < 3; j++ {
			buttonList[floor][dir] = false
		}
	}
}
