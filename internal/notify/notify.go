package notify

import (
	"fmt"
	"log"

	"github.com/gen2brain/beeep"
)

func Low(percent int) {
	go func() {
		err := beeep.Notify(
			"LAMZU Maya X: низкий заряд",
			fmt.Sprintf("Осталось %d%%. Подключите мышь к зарядке.", percent),
			"",
		)
		if err != nil {
			log.Println("уведомление:", err)
		}
	}()
}
