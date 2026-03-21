package membership

import (
	"context"
	"dpv/dpv/src/repository/graph"
	"log"
	"time"
)

// StartCron launches a background goroutine that periodically processes
// membership status transitions (e.g. approved -> active, cancelling -> cancelled).
func StartCron(db *graph.Db) {
	go func() {
		// Run ~1 min after start
		time.Sleep(1 * time.Minute)
		log.Println("Starting initial membership status transitions...")
		runTransitions(db)

		for {
			now := time.Now()
			// Next midnight
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 1, 0, 0, now.Location())
			waitDuration := next.Sub(now)
			time.Sleep(waitDuration)

			log.Println("Starting daily membership status transitions...")
			runTransitions(db)
		}
	}()
}

func runTransitions(db *graph.Db) {
	ctx := context.Background()
	nowUnix := time.Now().Unix()

	if err := db.ProcessMembershipTransitions(ctx, nowUnix); err != nil {
		log.Printf("Error processing membership transitions: %v\n", err)
	} else {
		log.Println("Membership status transitions processed successfully.")
	}
}
