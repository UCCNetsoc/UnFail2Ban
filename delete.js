(function() {

    document.addEventListener('DOMContentLoaded', init, false);
    
    var radio;
    var rows;
    var selected;
    var tableDiv; 
    var ipAddr;
    function init(){
        rows = document.querySelectorAll('.row');
        radio = document.querySelectorAll('#input')
        submit = document.getElementById("submit");
        tableDiv = document.getElementById("table");

        submit.addEventListener('click', submitted, false);
        
        for(var i = 0; i < input.length; i++){
            radio[i].addEventListener('change', ip, false);
            radio[i].selectedIndex = i;
        }
    }

    function ip(event){
        ipAddr = rows[event.target.selectedIndex].querySelectorAll('td')[4].innerHTML;
        console.log(ipAddr);
    }

    function submitted(event){
        event.preventDefault();
        request = new XMLHttpRequest();
        request.addEventListener('readystatechange', getResponse, false);
        request.open('GET', 'unban?ip='+ipAddr, true);
        request.send(null);
    }

    function getResponse(){
       if(request.readyState === 4){
            if(request.status === 200){
                tableDiv.innerHTML = request.responseText;
                window.alert("Success! "+ipAddr+" has been unbanned.")
            }else{
                window.alert("Uh oh! Something didnt go right :/")
            }
        }
    }
}());