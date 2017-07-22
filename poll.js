(function() {

    document.addEventListener('DOMContentLoaded', init, false);

    var logDiv;
    var date; 

    function init(){
        logDiv = document.getElementById("log");

        poll();
        setInterval(poll, 5000);
    }

    function poll(event){
        request = new XMLHttpRequest();
        request.onreadystatechange = getResponse;
        request.open('GET', 'poll?date="'+date+'"', true);
        request.send();
    }

    function getResponse(){
       if(request.readyState === 4){
            if(request.status === 200){  
                if(request.responseText != "") {
                    logDiv.innerHTML += request.responseText;
                    splitText = request.responseText.split("\n");
                    date = splitText[splitText.length-2].substr(0,23);
                    logDiv.scrollTop = logDiv.scrollHeight;
                }
            }
        }        
    }
}());