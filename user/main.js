htmx.onLoad((elt) => {
    let body = htmx.find("body");
    let gotifyToken = localStorage.getItem("gotify-login-key");
    if (gotifyToken === null || gotifyToken.trim().length == 0) {
        window.location = "/";
    }
    let display = htmx.find("#client-token-display");
    body.onload = async (test) => {
        display.textContent = localStorage.getItem("gotify-login-key");
    }
})