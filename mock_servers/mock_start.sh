#!/bin/bash
echo "Starting backend servers..."

# USERS SERVICES
go run backend_users1.go & 
go run backend_users2.go & 
go run backend_users3.go & 

# POSTS SERVICES
go run backend_posts1.go & 
go run backend_posts2.go & 

# Wait for all servers to start
sleep 3

# LOAD BALANCER
echo "Starting load balancer..."

