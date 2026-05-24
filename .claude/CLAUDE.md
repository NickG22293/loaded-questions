
We are building a web app that mimics the gameplay of the board game Loaded Questions. 

# Gameplay

To start, a person that is not part of a lobby will navigate to the website home page. There will be two buttons there - one for creating a Lobby and one for Joining one in progress. 

Starting a lobby brings them to the "Lobby" page where it lists the users who have joined. It assigns the lobby a UUID (short, like XYZ123), and that will be part of the URL. So something like, https://nick-gordon.com/questions/XYZ123. 

Joining a Lobby prompts the user to enter the UUID so they can be routed to that lobby page. 

The person who started the Lobby is the only one that can then Start the game from the lobby page. 

## Phase One - Question Asking

A group of up to 10 people are in a game, and each person takes their turn being the Asker. When you are the Asker, you enter a single question for the rest of the group to ask. The question is supposed to be introspective and have the Answerers thinking about themselves. Example questions: 

- If you could time travel back to one point in time for one year, what year would you go back to? 
- What is the one superpower you would have? 
- What's the one book you'd bring with you to a deserted island? 

The Asker submits the question to end Phase One. 

## Phase Two - Question Answering

When the question is posed, the Answerers have a limited amount of time to answer. Critically, when answers are submitted, they are ANONYMOUS to the Asker. When a user submits their answer, they wait for the rest of the users to submit. There should be an anonymous counter that counts the number of users out of the total that have submitted. When all users have submitted their answers, that ends Phase Two. 

## Phase Three - Answer Assignment 

Now the onus moves back to the Asker. The Asker reviews all the answers submitted. The goal of this phase is for the Asker to assign all the anonymous answers back to the correct person who answered them. They can either assign the answers as they first review them, or review them and then go back and assign them, or a combination. 

When an answer is assigned to a user, that answer pops up on their page along with the correct Answerer (if it isn't them). 

When all answers are assigned to users, the Asker has one last chance to review all the assigned answers and change them if necessary. Then they "Lock In" and that ends Phase Three. 

## Phase Four - Scoring 

Now the Answerer is notified of how many answers they correctly assigned, and which ones they missed. They get 1 point for every correctly assigned answer. Everyone laughs at the silly answers, and the next Asker is assigned. Then it reverts back to Phase 1. 

## Winning 

The winner is the first person to reach an arbitrary number set by the group at the start of the game. 

# Implementation 

The app employs a client/server architecture. 

## Frontend 

The frontend will use Tailwind + shadcn/ui for the component and styling library. It will use React. It'll be in Typescript. It should use whatever testing library is the industry standard for React and Typescript. 

It may be accessed over mobile or on a PC so we need to account for adjustable layout and sizing accordingly. 

## Backend 

The backend will be implemented in Go. We should leverage as much as we can - don't rewrite something that we can pull in from the standard library or a well-supported open source project. 

# Process 

Development and testing will move in stages. First serve the UI and server locally and poke about on localhost for rapid iteration, then a Dockerfile that hosts the UI via vite or an nginx server and the backend running in a simple Go binary, and then eventually in the Kubernetes cluster deployed via helm chart. 

# Docs 

The docs/ folder will hold additional specs that should be referenced for specific implementation details as needed. 