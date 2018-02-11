import * as $ from 'jquery'
    
$(document).ready(() => {
    $('.input').click(() => {
        let ipAddr  = $('.row')[$(this).parent().parent().index()].querySelectorAll('td')[4].innerHTML;
        let request = new XMLHttpRequest();
        request.onreadystatechange = () => {
            if(request.readyState === 4 && request.status === 200){  
                $('table').html(request.responseText)
                window.alert(ipAddr + ' has been unbanned.')
            }else{
                window.alert('Uh oh! Something didnt go right :/')
            }
        }
        request.open('GET', 'unban?ip='+ipAddr, true);
        request.send();
    })
})