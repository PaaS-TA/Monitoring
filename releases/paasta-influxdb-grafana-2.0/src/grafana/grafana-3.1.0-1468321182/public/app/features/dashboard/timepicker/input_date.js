/*! grafana - v3.1.0-1468321182 - 2016-07-12
 * Copyright (c) 2016 Torkel Ödegaard; Licensed Apache-2.0 */

System.register(["moment","app/core/utils/datemath"],function(a){function b(){return{restrict:"A",require:"ngModel",link:function(a,b,e,f){var g="YYYY-MM-DD HH:mm:ss",h=function(b){if(b.indexOf("now")!==-1)return d.isValid(b)?(f.$setValidity("error",!0),b):void f.$setValidity("error",!1);var e;return e=a.ctrl.isUtc?c["default"].utc(b,g):c["default"](b,g),e.isValid()?(f.$setValidity("error",!0),e):void f.$setValidity("error",!1)},i=function(a){return c["default"].isMoment(a)?a.format(g):a};f.$parsers.push(h),f.$formatters.push(i)}}}var c,d;return a("inputDateDirective",b),{setters:[function(a){c=a},function(a){d=a}],execute:function(){}}});