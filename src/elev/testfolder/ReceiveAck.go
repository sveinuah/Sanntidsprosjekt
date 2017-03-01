
func ReceiveAck(AckRxChan chan AckType, statusReqRxChan chan int, statusAckRxChan chan bool, buttonAckRxChan chan bool, executedOrdersAckRxChan chan bool, quitChan chan bool) {

	for{
		select {
		case quit <- quitChan:
			return
		case AckRec  =<- AckRxChan:
			if AckRec.Type == "Status" && AckRec.ID > 0{
				statusReqRxChan <- AckRec.ID
			}
			if AckRec.To == Name {

				switch AckRec.Type {
				case "Status":
					statusAckRxChan <- true
				case "ButtonPress":
					buttonAckRxChan <- true
				case "ExecOrder":
					executedOrdersAckRxChan <- true
				default:
				}
			}
		}
	}
}