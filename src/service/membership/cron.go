package membership

import (
	"context"
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/service/email"
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

	transitions, err := db.ProcessMembershipTransitions(ctx, nowUnix)
	if err != nil {
		log.Printf("Error processing membership transitions: %v\n", err)
		return
	}

	emailService := email.NewService(dpv.ConfigInstance)

	for _, u := range transitions.ActivatedUsers {
		u := u
		_ = emailService.SendMembershipBeganEmail(&u, nil)
	}
	for _, c := range transitions.ActivatedClubs {
		c := c
		_ = emailService.SendMembershipBeganEmail(nil, &c)
	}
	for _, u := range transitions.CancelledUsers {
		u := u
		_ = emailService.SendMembershipEndedEmail(&u, nil)
	}
	for _, c := range transitions.CancelledClubs {
		c := c
		_ = emailService.SendMembershipEndedEmail(nil, &c)
	}

	log.Println("Membership status transitions processed successfully.")
}

