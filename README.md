# 👵🏿 👴🏻 Bingo-box

Bingo box is an implementation of a bingo system, where the organizers can create and host a bingo game by printing id-assigned bingo cards and at the same time easily check if a given bingo card matches the wining patterns.

> Note: This system is only suited for 90-ball bingo, which is considered the "normal" kind of bingo in Denmark where i live 🇩🇰

## Motivation 🏃🏃

This is a project that I'm doing for fun, and at the same time it is a way for me to express my current level of Go understanding. But that is not the whole story. In my late highschool years, I was in charge for arranging and organizing a annual christmas bingo. I found it really frustrating to host the event, because of the immense time used on checking a potential winners bingo card pattern (one line, two lines and at last, the full plate). It was hard to hear every number shouted across the room, and maybe some number was not quite right, so then we would have to recite the whole pattern until we could confirm whether or not the player had won. Back then, in 2018, I built a bingo system to solve the problem, which is still in use to this day. But I have now desired to retire my old code and to develop a new system as the old codebase suffered from some hardcoding issues. This time with a backend built in Go instead of Node.js. The frontend will also get an overhaul, but will still be built in React as a single-page application.

## Project structure 🛠

### Frontend

The frontend is made in React, trying to adhere to the current best practices. The folder structure is made as flat as possible, to be as understandable as possible.

> Lives inside `/ui` folder.

### Backend

The project is structured with a domain-driven approach. All the data structures related to the domain will be in the root `bingo` package. All interactions with the domain will be made through interfaces and no dependencies will be made to to other places than the standard library from the root packge. While implementation details will be organized in packages according to their context. E.g `http` package is for everything related to the REST API.

Project Layout is inspired by Ben B. Johnsons way of structuring a Golang application. Read more about it [here](https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1) or look at the example source code application [here](https://github.com/benbjohnson/wtf)

> Lives inside `/server` folder.

## Running the app 🚀

As a natural first step to running this project, you have to clone this repository:

`git clone https://github.com/nohns/bingo-box.git`

For running the project, try one of the following methods

### Docker

For running all the components of this app the easiest way would be to use the Docker compose. Run the following command in the terminal of your choice:

`docker-compose up`

This will launch the both the server, database and the UI. Please go to `localhost:4000` to view the user interface

### Kubernetes (Helm chart)

Run following helm package
