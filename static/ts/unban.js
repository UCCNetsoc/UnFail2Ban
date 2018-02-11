"use strict";
var _this = this;
exports.__esModule = true;
var $ = require("jquery");
$(document).ready(function () {
    $('.input').click(function () {
        var ipAddr = $('.row')[$(_this).parent().parent().index()].querySelectorAll('td')[4].innerHTML;
        var request = new XMLHttpRequest();
        request.onreadystatechange = function () {
            if (request.readyState === 4 && request.status === 200) {
                $('table').html(request.responseText);
                window.alert(ipAddr + ' has been unbanned.');
            }
            else {
                window.alert('Uh oh! Something didnt go right :/');
            }
        };
        request.open('GET', 'unban?ip=' + ipAddr, true);
        request.send();
    });
});
