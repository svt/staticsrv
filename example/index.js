fetch("/config.json")
.then(response => response.json())
.then((payload) => {
    var user = payload.USER || "Unknown User"
    var origin = payload.ORIGIN || "JavaScript"

    var el = document.getElementById("user");
    el.innerText = user

    var el = document.getElementById("origin");
    el.innerText = origin
}).catch((error) => {
    console.log(error)
})

