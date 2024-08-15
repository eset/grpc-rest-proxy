// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package main

import (
	"context"
	"strings"
	"sync"

	pb "github.com/eset/grpc-rest-proxy/cmd/examples/grpcserver/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userService struct {
	pb.UnimplementedUserServiceServer
	users   []*pb.User
	summary *pb.Summary
	mx      sync.RWMutex
}

func NewUserService() *userService {
	return &userService{
		users:   users,
		summary: &pb.Summary{},
		mx:      sync.RWMutex{},
	}
}

func (s *userService) GetUser(ctx context.Context, request *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	for _, user := range s.users {
		if user.Username == request.Username {
			return &pb.GetUserResponse{
				User: user,
			}, nil
		}
	}

	errDetail := &pb.GetUserError{
		Username:       request.Username,
		Recommendation: "Please check the username and try again.",
	}

	errStatus, err := status.New(codes.NotFound, "User name not found.").WithDetails(errDetail)
	if err != nil {
		return nil, err
	}

	return nil, errStatus.Err()
}

func (s *userService) GetUsers(ctx context.Context, request *pb.GetUserRequest) (*pb.GetUsersResponse, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	var users []*pb.User
	for _, user := range s.users {
		if user.Username != request.Username {
			continue
		}
		if request.Country != "" {
			if user.Address.Country == request.Country {
				users = append(users, user)
			}
			continue
		}
		users = append(users, user)
	}
	return &pb.GetUsersResponse{
		Users: users,
	}, nil
}

func (s *userService) FilterUsers(ctx context.Context, request *pb.FilterUserRequest) (*pb.GetUsersResponse, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	var users []*pb.User
	for _, user := range s.users {
		if request.Username != "" {
			if request.Username != user.Username {
				continue
			}
		}
		if request.Country != "" {
			if request.Country != user.Address.Country {
				continue
			}
		}
		if request.Company != "" {
			if request.Company != user.Job.Company {
				continue
			}
		}
		if request.Jobtype != "" {
			if request.Jobtype != user.Job.JobType {
				continue
			}
		}
		users = append(users, user)
	}

	return &pb.GetUsersResponse{
		Users: users,
	}, nil
}

func (s *userService) CreateUser(ctx context.Context, request *pb.CreateUserRequest) (*pb.GetUserResponse, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	var newUserId int64 = 1
	if len(s.users) != 0 {
		lastUser := s.users[len(s.users)-1]
		newUserId = lastUser.Id + 1
	}
	s.users = append(s.users, &pb.User{
		Id:       newUserId,
		Username: request.User.Username,
		Surname:  request.User.Surname,
		Email:    request.User.Email,
		Job:      request.User.Job,
		Address:  request.User.Address,
	})
	return &pb.GetUserResponse{
		User: request.User,
	}, nil
}

func (s *userService) DeleteUser(ctx context.Context, request *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	for i, user := range s.users {
		if user.Username == request.Username {
			s.users = append(s.users[:i], s.users[i+1:]...)
			return &pb.DeleteUserResponse{
				Id: user.Id,
			}, nil
		}
	}
	return &pb.DeleteUserResponse{}, nil
}

func (s *userService) GetUsersByJobTitle(ctx context.Context, request *pb.GetUserRequest) (*pb.GetUsersResponse, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	var users []*pb.User
	for _, user := range s.users {
		if user.Job.JobTitle == request.Job.JobTitle {
			users = append(users, user)
		}
	}
	return &pb.GetUsersResponse{
		Users: users,
	}, nil
}

func (s *userService) GetUsersSummary(ctx context.Context, request *pb.GetSummaryRequest) (*pb.GetSummaryResponse, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	summary := strings.Split(request.Summary, "/")
	for _, kind := range summary {
		switch kind {
		case "username":
			for _, user := range s.users {
				s.summary.Usernames = append(s.summary.Usernames, user.Username)
			}
		case "country":
			for _, country := range s.users {
				s.summary.Countries = append(s.summary.Countries, country.Address.Country)
			}
		case "jobtitle":
			for _, jobTitle := range s.users {
				s.summary.Countries = append(s.summary.JobTitles, jobTitle.Job.JobTitle)
			}
		case "jobtype":
			for _, jobType := range s.users {
				s.summary.Countries = append(s.summary.JobTitles, jobType.Job.JobType)
			}
		}
	}
	return &pb.GetSummaryResponse{
		Summary: &pb.Summary{
			Usernames: s.summary.Usernames,
			Countries: s.summary.Countries,
			JobTitles: s.summary.JobTitles,
			JobTypes:  s.summary.JobTypes,
		},
	}, nil
}

func (s *userService) UpdateUserJob(ctx context.Context, request *pb.UpdateUserJobRequest) (*pb.UpdateUserJobResponse, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, user := range s.users {
		if user.Username == request.Username {
			user.Job.JobTitle = request.Job.JobTitle
			return &pb.UpdateUserJobResponse{
				User: user,
			}, nil
		}
	}
	return &pb.UpdateUserJobResponse{}, nil
}

func (s *userService) GetUsersPost(ctx context.Context, request *pb.GetUserPostRequest) (*pb.GetUserPostResponse, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	var users []*pb.User
	for _, user := range s.users {
		if user.Address.Country != request.Address.Country {
			continue
		}
		if user.Post.String() != strings.ToUpper(request.Type) {
			continue
		}
		users = append(users, user)

		return &pb.GetUserPostResponse{
			Users: users,
		}, nil
	}
	return &pb.GetUserPostResponse{}, nil
}

var users = []*pb.User{
	{
		Id:       1,
		Username: "Sylvia",
		Surname:  "Stanton",
		Email:    "Sylvia_Stanton9@example.org",
		Address: &pb.Address{
			City:        "Volkmanside",
			Country:     "Antarctica",
			CountryCode: "UA",
		},
		Job: &pb.Job{
			Company:  "Labadie - Heaney",
			JobArea:  "Assurance",
			JobTitle: "Central Solutions Engineer",
			JobType:  "Coordinator",
		},
		Post: pb.Post_COMPETITION,
	},
	{
		Id:       2,
		Username: "Sylvia",
		Surname:  "Murray",
		Email:    "Sylvia_Murray50@example.org",
		Address: &pb.Address{
			City:        "Bradleyfurt",
			Country:     "Bahrain",
			CountryCode: "PM",
		},
		Job: &pb.Job{
			Company:  "Rath Group",
			JobArea:  "Group",
			JobTitle: "Central Applications Agent",
			JobType:  "Analyst",
		},
		Post: pb.Post_ENGAGEMENT,
	},
	{
		Id:       3,
		Username: "Rita",
		Surname:  "Schuster",
		Email:    "Rita8@example.org",
		Address: &pb.Address{
			City:        "Yeseniafurt",
			Country:     "Mauritania",
			CountryCode: "MN",
		},
		Job: &pb.Job{
			Company:  "Casper, Hackett and Rath",
			JobArea:  "Group",
			JobTitle: "Corporate Data Strategist",
			JobType:  "Executive",
		},
		Post: pb.Post_PRODUCT,
	},
	{
		Id:       4,
		Username: "Bryant",
		Surname:  "Stehr",
		Email:    "Bryant.Stehr56@example.org",
		Address: &pb.Address{
			City:        "North Morganton",
			Country:     "Zambia",
			CountryCode: "BT",
		},
		Job: &pb.Job{
			Company:  "Legros, Adams and Hilll",
			JobArea:  "Brand",
			JobTitle: "Investor Accounts Orchestrator",
			JobType:  "Manager",
		},
		Post: pb.Post_NEWS_TRENDING,
	},
	{
		Id:       5,
		Username: "Hattie",
		Surname:  "Stamm",
		Email:    "Hattie_Stamm11@example.org",
		Address: &pb.Address{
			City:        "Lake Tillman",
			Country:     "Vanuatu",
			CountryCode: "YE",
		},
		Job: &pb.Job{
			Company:  "Hartmann, Nienow and Swaniawski",
			JobArea:  "Integration",
			JobTitle: "Central Implementation Engineer",
			JobType:  "Engineer",
		},
		Post: pb.Post_PROMOTION,
	},
	{
		Id:       6,
		Username: "Marta",
		Surname:  "Buckridge",
		Email:    "Marta_Buckridge@example.org",
		Address: &pb.Address{
			City:        "Eliezerburgh",
			Country:     "Iraq",
			CountryCode: "AL",
		},
		Job: &pb.Job{
			Company:  "Kuphal and Sons",
			JobArea:  "Security",
			JobTitle: "Investor Assurance Strategist",
			JobType:  "Producer",
		},
		Post: pb.Post_NEWS_TRENDING,
	},
	{
		Id:       7,
		Username: "Kelley",
		Surname:  "Stoltenberg",
		Email:    "Kelley.Stoltenberg@example.org",
		Address: &pb.Address{
			City:        "New Lawrenceboro",
			Country:     "Isle of",
			CountryCode: "AW",
		},
		Job: &pb.Job{
			Company:  "Bartell, Kerluke and Koepp",
			JobArea:  "Solutions",
			JobTitle: "Architect",
			JobType:  "Assistant",
		},
		Post: pb.Post_COMPETITION,
	},
	{
		Id:       8,
		Username: "John",
		Surname:  "Doe",
		Email:    "john.doe@example.com",
		Address: &pb.Address{
			City:        "New York",
			Country:     "USA",
			CountryCode: "US",
		},
		Job: &pb.Job{
			Company:  "ABC Company",
			JobArea:  "IT",
			JobTitle: "Software Engineer",
			JobType:  "Full-time",
		},
		Post: pb.Post_ENGAGEMENT,
	},
	{
		Id:       9,
		Username: "Jane",
		Surname:  "Smith",
		Email:    "jane.smith@example.com",
		Address: &pb.Address{
			City:        "San Francisco",
			Country:     "USA",
			CountryCode: "US",
		},
		Job: &pb.Job{
			Company:  "XYZ Corporation",
			JobArea:  "Finance",
			JobTitle: "Financial Analyst",
			JobType:  "Part-time",
		},
		Post: pb.Post_PRODUCT,
	},
	{
		Id:       10,
		Username: "Mike",
		Surname:  "Johnson",
		Email:    "mike.johnson@example.com",
		Address: &pb.Address{
			City:        "London",
			Country:     "UK",
			CountryCode: "GB",
		},
		Job: &pb.Job{
			Company:  "DEF Ltd",
			JobArea:  "Marketing",
			JobTitle: "Marketing Manager",
			JobType:  "Contract",
		},
		Post: pb.Post_PROMOTION,
	},
	{
		Id:       11,
		Username: "Alice",
		Surname:  "Brown",
		Email:    "alice.brown@example.com",
		Address: &pb.Address{
			City:        "Paris",
			Country:     "France",
			CountryCode: "FR",
		},
		Job: &pb.Job{
			Company:  "GHI Corporation",
			JobArea:  "Sales",
			JobTitle: "Sales Representative",
			JobType:  "Full-time",
		},
		Post: pb.Post_NEWS_TRENDING,
	},
	{
		Id:       12,
		Username: "David",
		Surname:  "Wilson",
		Email:    "david.wilson@example.com",
		Address: &pb.Address{
			City:        "Berlin",
			Country:     "Germany",
			CountryCode: "DE",
		},
		Job: &pb.Job{
			Company:  "JKL Company",
			JobArea:  "Engineering",
			JobTitle: "Software Developer",
			JobType:  "Full-time",
		},
		Post: pb.Post_PROMOTION,
	},
	{
		Id:       13,
		Username: "Emily",
		Surname:  "Taylor",
		Email:    "emily.taylor@example.com",
		Address: &pb.Address{
			City:        "Tokyo",
			Country:     "Japan",
			CountryCode: "JP",
		},
		Job: &pb.Job{
			Company:  "MNO Corporation",
			JobArea:  "Marketing",
			JobTitle: "Marketing Specialist",
			JobType:  "Part-time",
		},
		Post: pb.Post_COMPETITION,
	},
	{
		Id:       14,
		Username: "Michael",
		Surname:  "Anderson",
		Email:    "michael.anderson@example.com",
		Address: &pb.Address{
			City:        "Sydney",
			Country:     "Australia",
			CountryCode: "AU",
		},
		Job: &pb.Job{
			Company:  "PQR Ltd",
			JobArea:  "Finance",
			JobTitle: "Financial Advisor",
			JobType:  "Contract",
		},
		Post: pb.Post_ENGAGEMENT,
	},
	{
		Id:       15,
		Username: "Olivia",
		Surname:  "Clark",
		Email:    "olivia.clark@example.com",
		Address: &pb.Address{
			City:        "Toronto",
			Country:     "Canada",
			CountryCode: "CA",
		},
		Job: &pb.Job{
			Company:  "STU Company",
			JobArea:  "IT",
			JobTitle: "System Administrator",
			JobType:  "Full-time",
		},
		Post: pb.Post_PRODUCT,
	},
	{
		Id:       16,
		Username: "William",
		Surname:  "Walker",
		Email:    "william.walker@example.com",
		Address: &pb.Address{
			City:        "Moscow",
			Country:     "Russia",
			CountryCode: "RU",
		},
		Job: &pb.Job{
			Company:  "VWX Corporation",
			JobArea:  "Engineering",
			JobTitle: "Hardware Engineer",
			JobType:  "Full-time",
		},
		Post: pb.Post_NEWS_TRENDING,
	},
	{
		Id:       17,
		Username: "Sophia",
		Surname:  "Lewis",
		Email:    "sophia.lewis@example.com",
		Address: &pb.Address{
			City:        "Rome",
			Country:     "Italy",
			CountryCode: "IT",
		},
		Job: &pb.Job{
			Company:  "YZ Company",
			JobArea:  "Sales",
			JobTitle: "Sales Manager",
			JobType:  "Full-time",
		},
		Post: pb.Post_PROMOTION,
	},
	{
		Id:       18,
		Username: "James",
		Surname:  "Harris",
		Email:    "james.harris@example.com",
		Address: &pb.Address{
			City:        "Madrid",
			Country:     "Spain",
			CountryCode: "ES",
		},
		Job: &pb.Job{
			Company:  "ABC Company",
			JobArea:  "Marketing",
			JobTitle: "Marketing Coordinator",
			JobType:  "Part-time",
		},
		Post: pb.Post_COMPETITION,
	},
	{
		Id:       19,
		Username: "Ava",
		Surname:  "Young",
		Email:    "ava.young@example.com",
		Address: &pb.Address{
			City:        "Seoul",
			Country:     "South Korea",
			CountryCode: "KR",
		},
		Job: &pb.Job{
			Company:  "DEF Ltd",
			JobArea:  "Finance",
			JobTitle: "Financial Planner",
			JobType:  "Contract",
		},
		Post: pb.Post_ENGAGEMENT,
	},
	{
		Id:       20,
		Username: "Logan",
		Surname:  "Lee",
		Email:    "logan.lee@example.com",
		Address: &pb.Address{
			City:        "Beijing",
			Country:     "China",
			CountryCode: "CN",
		},
		Job: &pb.Job{
			Company:  "GHI Corporation",
			JobArea:  "IT",
			JobTitle: "Software Developer",
			JobType:  "Full-time",
		},
		Post: pb.Post_PRODUCT,
	},
	{
		Id:       21,
		Username: "Mia",
		Surname:  "Hall",
		Email:    "mia.hall@example.com",
		Address: &pb.Address{
			City:        "Mexico City",
			Country:     "Mexico",
			CountryCode: "MX",
		},
		Job: &pb.Job{
			Company:  "JKL Company",
			JobArea:  "Engineering",
			JobTitle: "Hardware Engineer",
			JobType:  "Full-time",
		},
	},
	{
		Id:       22,
		Username: "Benjamin",
		Surname:  "Young",
		Email:    "benjamin.young@example.com",
		Address: &pb.Address{
			City:        "Cairo",
			Country:     "Egypt",
			CountryCode: "EG",
		},
		Job: &pb.Job{
			Company:  "MNO Corporation",
			JobArea:  "Sales",
			JobTitle: "Sales Representative",
			JobType:  "Full-time",
		},
	},
	{
		Id:       23,
		Username: "Charlotte",
		Surname:  "King",
		Email:    "charlotte.king@example.com",
		Address: &pb.Address{
			City:        "Sydney",
			Country:     "Australia",
			CountryCode: "AU",
		},
		Job: &pb.Job{
			Company:  "PQR Ltd",
			JobArea:  "Marketing",
			JobTitle: "Marketing Specialist",
			JobType:  "Part-time",
		},
	},
	{
		Id:       24,
		Username: "Henry",
		Surname:  "Wright",
		Email:    "henry.wright@example.com",
		Address: &pb.Address{
			City:        "Toronto",
			Country:     "Canada",
			CountryCode: "CA",
		},
		Job: &pb.Job{
			Company:  "STU Company",
			JobArea:  "IT",
			JobTitle: "System Administrator",
			JobType:  "Full-time",
		},
	},
	{
		Id:       25,
		Username: "Amelia",
		Surname:  "Lopez",
		Email:    "amelia.lopez@example.com",
		Address: &pb.Address{
			City:        "Moscow",
			Country:     "Russia",
			CountryCode: "RU",
		},
		Job: &pb.Job{
			Company:  "VWX Corporation",
			JobArea:  "Engineering",
			JobTitle: "Hardware Engineer",
			JobType:  "Full-time",
		},
	},
}
