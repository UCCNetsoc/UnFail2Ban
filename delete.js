(function() {

    document.addEventListener('DOMContentLoaded', init, false);
    
    var rows;
    var tableDiv; 
    var ipAddr;

    function init(){
        rows = document.querySelectorAll('.row');
        var buttons = document.querySelectorAll('.input')
        tableDiv = document.getElementById("table");
        
        for(var i = 0; i < buttons.length; i++){
            buttons[i].addEventListener('click', submitted);
            buttons[i].selectedIndex = i;
        }
    }

    function submitted(event){
        ipAddr = rows[event.target.selectedIndex].querySelectorAll('td')[4].innerHTML;
        request = new XMLHttpRequest();
        request.onreadystatechange = getResponse;
        request.open('GET', 'unban?ip='+ipAddr, true);
        request.send();
    }

    function getResponse(){
       if(request.readyState === 4){
            if(request.status === 200){  
                tableDiv.innerHTML = request.responseText;
                init()
                window.alert(ipAddr  +" has been unbanned.")
            }else{
                window.alert("Uh oh! Something didnt go right :/")
            }
        }
    }
}());