# SplitWise API
This application provides a RESTful API for managing splits among friends or family, allowing users to efficiently track and manage shared expenses.

# Prerequisites
Ensure Go is installed your machine.
PostgreSQL installed and running.

# Database Setup
Run the provided SQL scripts to set up the database

# Configuration
In the Go code, update the database connection details (URL, path, username, and password) as needed.

# Running the Application
1.Clone the repository:
git clone <repository-url>

2.Navigate to the project directory:
cd SplitWiseAPI

3.Run go mod tidy to clean up the dependencies:
go mod tidy

4.Build the application:
go build

5.Start the application:
./SplitWiseAPI
