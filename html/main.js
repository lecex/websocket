window.addEventListener("load", function (evt) {
    var wsUri = "ws://127.0.0.1:8082/ws"
    var output = document.getElementById("output");
    var nameTxt = document.getElementById("name");
    var ws;
    var print = function (message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };

    (function () {
        ws = new WebSocket(wsUri);

        ws.onopen = function (evt) {
            print('<span style="color: green;">Connection Open</span>');

            var info = {
                deviceInfo:"TXAP12107000994ND002113",
                service:'user-api',
                method:'Auth.Auth',
                request:{
                    user: {
                        username: "admin", 
                        password: "admin123"
                    },
                },
            }
            ws.send(JSON.stringify(info));
        }
        ws.onclose = function (evt) {
            print('<span style="color: red;">Connection Closed</span>');
            ws = null;
        }
        ws.onmessage = function (evt) {
            const a= JSON.parse(evt.data)
            console.log(a);
            print('<span style="color: blue;">Update: </span>' + evt.data);
        }
        ws.onerror = function (evt) {
            print('<span style="color: red;">Error: </span>' + evt.data);
        }
    })();


    document.getElementById("send").onclick = function (evt) {
        if (!ws) {
            return false
        }

        var msg = {hi: nameTxt.value}

        req = JSON.stringify(msg)
        print('<span style="color: blue;">Sent request: </span>' + req);
        ws.send(JSON.stringify(msg));

        return false;
    };

    document.getElementById("cancel").onclick = function (evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        print('<span style="color: red;">Request Canceled</span>');
        return false;
    };

    document.getElementById("open").onclick = function (evt) {
        if (!ws) {
            newSocket()
        }
        return false;
    };
})
