db = db.getSiblingDB("confa")
if (db.getUser("confa") === null) {
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
}

db = db.getSiblingDB("iam")
if (db.getUser("iam") === null) {
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
}