function getHttpRequest() {
    var xmlhttp;
    if (window.XMLHttpRequest) {// code for IE7+, Firefox, Chrome, Opera, Safari
        xmlhttp = new XMLHttpRequest();
    } else {// code for IE6, IE5
        xmlhttp = new ActiveXObject("Microsoft.XMLHTTP");
    }

    return xmlhttp;
}

function sendAJAX(path, params, callback) {
    var method = "POST";
    var url = path + "?lrt=" + (new Date().getTime());
    
    var str = [];

    for (var key in params) {
        if (params.hasOwnProperty(key)) {
            str.push(encodeURIComponent(key) + "=" + encodeURIComponent(params[key]));
        }
    }

    var xhr = getHttpRequest();
    xhr.open(method, url, true);

    xhr.onreadystatechange = function() {
        if (this.readyState == 4 && this.status == 200) {
            if (callback != null && callback != undefined) {
                callback(this.responseText);
            }
        }
    }

    xhr.setRequestHeader("Content-type","application/x-www-form-urlencoded");
    xhr.send(str.join("&"));
}

function sendPOST(path, params) {
    var method = "post";
    var url = path + "?lrt=" + (new Date().getTime());

    // The rest of this code assumes you are not using a library.
    // It can be made less wordy if you use one.
    var form = document.createElement("form");
    form.target = '_blank';
    form.setAttribute("method", method);
    form.setAttribute("action", url);

    for(var key in params) {
        if(params.hasOwnProperty(key)) {
            var hiddenField = document.createElement("input");
            hiddenField.setAttribute("type", "hidden");
            hiddenField.setAttribute("name", key);
            hiddenField.setAttribute("value", params[key]);

            form.appendChild(hiddenField);
         }
    }

    document.body.appendChild(form);
    form.submit();
    document.body.removeChild(form);
}
