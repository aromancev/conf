db = db.getSiblingDB("confa")
db.createUser({
    user: "confa",
    pwd: confaPwd, // This should be evaluated with --eval when running this script.
    roles: [
        {
            role: "readWrite",
            db: "confa"
        },
    ]
})

db = db.getSiblingDB("iam")
db.createUser({
    user: "iam",
    pwd: iamPwd, // This should be evaluated with --eval when running this script.
    roles: [
        {
            role: "readWrite",
            db: "iam"
        },
    ]
})
