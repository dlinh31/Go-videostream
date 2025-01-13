package server

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/dlinh31/go-videostream/proto"
)

var (
    watchParties     = make(map[string][]string)
    mu               sync.Mutex
)

func (*VideoStreamServer) CreateWatchParty(ctx context.Context ,req *pb.CreatePartyRequest) (*pb.PartyResponse, error){
    mu.Lock()
    defer mu.Unlock()
    partyId := fmt.Sprintf("party-%d", len(watchParties) + 1)
    watchParties[partyId] = []string{req.HostName}
    return &pb.PartyResponse{PartyId: partyId, Status: "success",Users: watchParties[partyId]}, nil
}

func (*VideoStreamServer) JoinWatchParty(ctx context.Context, req *pb.JoinPartyRequest) (*pb.PartyResponse, error){
    mu.Lock()
    defer mu.Unlock()
    users, exists := watchParties[req.PartyId];
    if !exists{
        return &pb.PartyResponse{PartyId: req.PartyId, Status: "party not found", Users: nil}, nil
    }
    watchParties[req.PartyId] = append(users, req.UserName)
    return &pb.PartyResponse{PartyId: req.PartyId, Status: "success", Users: watchParties[req.PartyId]}, nil
}