package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pbUsers "github.com/tsingson/grpc-postgres/proto"
)

var (
	addr      = flag.String("addr", ":10000", "The address of the gRPC server")
	cert      = flag.String("cert", "../insecure/cert.pem", "The path of the server certificate")
	olderThan = flag.Duration("older_than", 0, "Filter to use when listing users.")
	add       = flag.String("add", "test", "Whether to add another user")
)

func main() {
	flag.Parse()
	if len(*add) > 0 {
		fmt.Println("----------------> user name:", *add)
	}

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	}

	// creds, err := credentials.NewClientTLSFromFile(*cert, "")
	// if err != nil {
	// 	log.WithError(err).Fatal("Failed to create server credentials")
	// }

	conn, err := grpc.Dial(
		*addr,
		grpc.WithInsecure(),
		// grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to dial the server")
	}

	c := pbUsers.NewUserServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(*add) > 0 {
		user, err := c.AddUser(ctx, &pbUsers.AddUserRequest{
			Role:     pbUsers.Role_GUEST,
			UserName: *add,
		})
		if err != nil {
			log.WithError(err).Fatal("Failed to add user")
		}

		t, err := ptypes.Timestamp(user.GetCreateTime())
		if err != nil {
			log.WithError(err).Error("Failed to list users")
		}

		log.WithFields(logrus.Fields{
			"id":          user.GetId(),
			"user_name":   user.GetUserName(),
			"role":        user.GetRole().String(),
			"create_time": t.Local().Format(time.RFC3339),
		}).Info("Added user")
	}

	lReq := new(pbUsers.ListUsersRequest)

	if *olderThan != 0 {
		lReq.OlderThan = ptypes.DurationProto(*olderThan)
	}

	srv, err := c.ListUsers(ctx, lReq)
	if err != nil {
		log.WithError(err).Fatal("Failed to list users")
	}

	for {
		user, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.WithError(err).Fatal("Error while receiving users")
		}

		t, err := ptypes.Timestamp(user.GetCreateTime())
		if err != nil {
			log.WithError(err).Error("Failed to list users")
		}

		log.WithFields(logrus.Fields{
			"id":          user.GetId(),
			"user_name":   user.GetUserName(),
			"role":        user.GetRole().String(),
			"create_time": t.Local().Format(time.RFC3339),
		}).Info("Read user")
	}

	log.Info("Finished")
}
