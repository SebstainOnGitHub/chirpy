#HOW THE JWT WORKS

POST /api/users 
with a body like:
{
    "email": "saul@bettercall.com",
    "password": 123456
}
Creates a user. Now there's one in the database!

POST /api/login

with a body like:
{
    "email": "saul@bettercall.com",
    "password": "123456"
}
Logs a user in. Now theres a logged in user!
This will get a response like
{
    same email,
    same password,
    "token": ${long ahh token}
}
Now the computer has a way to identify you as a requester person, as otherwise how will it know who wants to replace their password!?

so finallyy

PUT /api/users

with a header like:
{
    //The bearer is just used to signify the person making the request (who BEARS the token)
    "Authorization": "Bearer ${long ahh token}"
}
and a body like this
{
    "email": a different email,
    "password": a different password,
}

sucessfully updates a user's information, as the computer knows whos making the request 






ACCESS TOKENS

Access tokens do not allow you to log in immediately, they allow you to be able to get JWT's immediately which then allow you to log in

