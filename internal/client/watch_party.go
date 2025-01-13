package client

import (
	"context"
	"log"
	"time"
	pb "github.com/dlinh31/go-videostream/proto"

)




func (vc *VideoClient) CreateWatchParty(hostName string) (string, error){
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	res, err := vc.client.CreateWatchParty(ctx, &pb.CreatePartyRequest{HostName: hostName})
	if err != nil {
		return "", err
	}
	log.Printf("Created watch party: %s, Participants: %v", res.PartyId, res.Users)
	return res.PartyId, nil
}

func (vc *VideoClient) JoinWatchParty(partyId, userName string) (error){
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	res, err := vc.client.JoinWatchParty(ctx, &pb.JoinPartyRequest{PartyId: partyId, UserName: userName})
	if err != nil {
		return err
	}
	log.Printf("Joined watch party: %s, Participants: %v", res.PartyId, res.Users)
	return nil
}

