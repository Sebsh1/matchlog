### A service for handling you in-office foosball needs
Does your office have a foosball table? Do you employ software developers?
Then they have no doubt discussed coming up with a rating system for their matches to see how they stack up against each other. \
This service will save their development time for more pressing features with *actual* deadlines.

This repository has a complete backend and REST API for logging matches to a database and calculating rating, leaderboards and player stats. \
All you have to do is deploy it, play foosball and log the results. 

**Notice:** Your developers might instead start thinking about creating a frontend for this service to throw onto a screen near the table.

### Setup
Run <code>go run main.go serve</code> to start the service with the config defined in <code>config.yaml</code>.

If you wish to iterate on this service, running <code>go run main.go seed</code> creates some dummy data in the database. This command should not be run on your "production" database.

### API Docs

### TODO
- DELETE player/:id needs to remove player from table players_teams
- GET player/:id/stats needs to be implemented
- POST match fails to update player ratings
- Authguard on endpoints?