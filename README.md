# chorus

Step 1 - Create the MVP which is a single server, no other tools that allows users to connect and get put in a lobby that is managed by a  javascript.  That script can then do matchmaking to put you in a tic-tac-toe game, a room managed by another javascript.

This works well enough.

Step 2 - Route everything through RedPanda/Kafka.  How does this change things

A user connects to the server.  It is given rooms to listen to.  It connects to that rooms topic and filters based on criteria.