!function(){function t(r){return t="function"==typeof Symbol&&"symbol"==typeof Symbol.iterator?function(t){return typeof t}:function(t){return t&&"function"==typeof Symbol&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},t(r)}function r(){"use strict";/*! regenerator-runtime -- Copyright (c) 2014-present, Facebook, Inc. -- license (MIT): https://github.com/facebook/regenerator/blob/main/LICENSE */r=function(){return e};var e={},n=Object.prototype,o=n.hasOwnProperty,i="function"==typeof Symbol?Symbol:{},a=i.iterator||"@@iterator",c=i.asyncIterator||"@@asyncIterator",u=i.toStringTag||"@@toStringTag";function l(t,r,e){return Object.defineProperty(t,r,{value:e,enumerable:!0,configurable:!0,writable:!0}),t[r]}try{l({},"")}catch(j){l=function(t,r,e){return t[r]=e}}function s(t,r,e,n){var o=r&&r.prototype instanceof d?r:d,i=Object.create(o.prototype),a=new E(n||[]);return i._invoke=function(t,r,e){var n="suspendedStart";return function(o,i){if("executing"===n)throw new Error("Generator is already running");if("completed"===n){if("throw"===o)throw i;return _()}for(e.method=o,e.arg=i;;){var a=e.delegate;if(a){var c=k(a,e);if(c){if(c===h)continue;return c}}if("next"===e.method)e.sent=e._sent=e.arg;else if("throw"===e.method){if("suspendedStart"===n)throw n="completed",e.arg;e.dispatchException(e.arg)}else"return"===e.method&&e.abrupt("return",e.arg);n="executing";var u=f(t,r,e);if("normal"===u.type){if(n=e.done?"completed":"suspendedYield",u.arg===h)continue;return{value:u.arg,done:e.done}}"throw"===u.type&&(n="completed",e.method="throw",e.arg=u.arg)}}}(t,e,a),i}function f(t,r,e){try{return{type:"normal",arg:t.call(r,e)}}catch(j){return{type:"throw",arg:j}}}e.wrap=s;var h={};function d(){}function p(){}function g(){}var y={};l(y,a,(function(){return this}));var v=Object.getPrototypeOf,m=v&&v(v(S([])));m&&m!==n&&o.call(m,a)&&(y=m);var w=g.prototype=d.prototype=Object.create(y);function b(t){["next","throw","return"].forEach((function(r){l(t,r,(function(t){return this._invoke(r,t)}))}))}function x(r,e){function n(i,a,c,u){var l=f(r[i],r,a);if("throw"!==l.type){var s=l.arg,h=s.value;return h&&"object"==t(h)&&o.call(h,"__await")?e.resolve(h.__await).then((function(t){n("next",t,c,u)}),(function(t){n("throw",t,c,u)})):e.resolve(h).then((function(t){s.value=t,c(s)}),(function(t){return n("throw",t,c,u)}))}u(l.arg)}var i;this._invoke=function(t,r){function o(){return new e((function(e,o){n(t,r,e,o)}))}return i=i?i.then(o,o):o()}}function k(t,r){var e=t.iterator[r.method];if(void 0===e){if(r.delegate=null,"throw"===r.method){if(t.iterator.return&&(r.method="return",r.arg=void 0,k(t,r),"throw"===r.method))return h;r.method="throw",r.arg=new TypeError("The iterator does not provide a 'throw' method")}return h}var n=f(e,t.iterator,r.arg);if("throw"===n.type)return r.method="throw",r.arg=n.arg,r.delegate=null,h;var o=n.arg;return o?o.done?(r[t.resultName]=o.value,r.next=t.nextLoc,"return"!==r.method&&(r.method="next",r.arg=void 0),r.delegate=null,h):o:(r.method="throw",r.arg=new TypeError("iterator result is not an object"),r.delegate=null,h)}function L(t){var r={tryLoc:t[0]};1 in t&&(r.catchLoc=t[1]),2 in t&&(r.finallyLoc=t[2],r.afterLoc=t[3]),this.tryEntries.push(r)}function $(t){var r=t.completion||{};r.type="normal",delete r.arg,t.completion=r}function E(t){this.tryEntries=[{tryLoc:"root"}],t.forEach(L,this),this.reset(!0)}function S(t){if(t){var r=t[a];if(r)return r.call(t);if("function"==typeof t.next)return t;if(!isNaN(t.length)){var e=-1,n=function r(){for(;++e<t.length;)if(o.call(t,e))return r.value=t[e],r.done=!1,r;return r.value=void 0,r.done=!0,r};return n.next=n}}return{next:_}}function _(){return{value:void 0,done:!0}}return p.prototype=g,l(w,"constructor",g),l(g,"constructor",p),p.displayName=l(g,u,"GeneratorFunction"),e.isGeneratorFunction=function(t){var r="function"==typeof t&&t.constructor;return!!r&&(r===p||"GeneratorFunction"===(r.displayName||r.name))},e.mark=function(t){return Object.setPrototypeOf?Object.setPrototypeOf(t,g):(t.__proto__=g,l(t,u,"GeneratorFunction")),t.prototype=Object.create(w),t},e.awrap=function(t){return{__await:t}},b(x.prototype),l(x.prototype,c,(function(){return this})),e.AsyncIterator=x,e.async=function(t,r,n,o,i){void 0===i&&(i=Promise);var a=new x(s(t,r,n,o),i);return e.isGeneratorFunction(r)?a:a.next().then((function(t){return t.done?t.value:a.next()}))},b(w),l(w,u,"Generator"),l(w,a,(function(){return this})),l(w,"toString",(function(){return"[object Generator]"})),e.keys=function(t){var r=[];for(var e in t)r.push(e);return r.reverse(),function e(){for(;r.length;){var n=r.pop();if(n in t)return e.value=n,e.done=!1,e}return e.done=!0,e}},e.values=S,E.prototype={constructor:E,reset:function(t){if(this.prev=0,this.next=0,this.sent=this._sent=void 0,this.done=!1,this.delegate=null,this.method="next",this.arg=void 0,this.tryEntries.forEach($),!t)for(var r in this)"t"===r.charAt(0)&&o.call(this,r)&&!isNaN(+r.slice(1))&&(this[r]=void 0)},stop:function(){this.done=!0;var t=this.tryEntries[0].completion;if("throw"===t.type)throw t.arg;return this.rval},dispatchException:function(t){if(this.done)throw t;var r=this;function e(e,n){return a.type="throw",a.arg=t,r.next=e,n&&(r.method="next",r.arg=void 0),!!n}for(var n=this.tryEntries.length-1;n>=0;--n){var i=this.tryEntries[n],a=i.completion;if("root"===i.tryLoc)return e("end");if(i.tryLoc<=this.prev){var c=o.call(i,"catchLoc"),u=o.call(i,"finallyLoc");if(c&&u){if(this.prev<i.catchLoc)return e(i.catchLoc,!0);if(this.prev<i.finallyLoc)return e(i.finallyLoc)}else if(c){if(this.prev<i.catchLoc)return e(i.catchLoc,!0)}else{if(!u)throw new Error("try statement without catch or finally");if(this.prev<i.finallyLoc)return e(i.finallyLoc)}}}},abrupt:function(t,r){for(var e=this.tryEntries.length-1;e>=0;--e){var n=this.tryEntries[e];if(n.tryLoc<=this.prev&&o.call(n,"finallyLoc")&&this.prev<n.finallyLoc){var i=n;break}}i&&("break"===t||"continue"===t)&&i.tryLoc<=r&&r<=i.finallyLoc&&(i=null);var a=i?i.completion:{};return a.type=t,a.arg=r,i?(this.method="next",this.next=i.finallyLoc,h):this.complete(a)},complete:function(t,r){if("throw"===t.type)throw t.arg;return"break"===t.type||"continue"===t.type?this.next=t.arg:"return"===t.type?(this.rval=this.arg=t.arg,this.method="return",this.next="end"):"normal"===t.type&&r&&(this.next=r),h},finish:function(t){for(var r=this.tryEntries.length-1;r>=0;--r){var e=this.tryEntries[r];if(e.finallyLoc===t)return this.complete(e.completion,e.afterLoc),$(e),h}},catch:function(t){for(var r=this.tryEntries.length-1;r>=0;--r){var e=this.tryEntries[r];if(e.tryLoc===t){var n=e.completion;if("throw"===n.type){var o=n.arg;$(e)}return o}}throw new Error("illegal catch attempt")},delegateYield:function(t,r,e){return this.delegate={iterator:S(t),resultName:r,nextLoc:e},"next"===this.method&&(this.arg=void 0),h}},e}function e(t,r,e,n,o,i,a){try{var c=t[i](a),u=c.value}catch(l){return void e(l)}c.done?r(u):Promise.resolve(u).then(n,o)}function n(t){return function(){var r=this,n=arguments;return new Promise((function(o,i){var a=t.apply(r,n);function c(t){e(a,o,i,c,u,"next",t)}function u(t){e(a,o,i,c,u,"throw",t)}c(void 0)}))}}function o(t,r){return function(t){if(Array.isArray(t))return t}(t)||function(t,r){var e=null==t?null:"undefined"!=typeof Symbol&&t[Symbol.iterator]||t["@@iterator"];if(null==e)return;var n,o,i=[],a=!0,c=!1;try{for(e=e.call(t);!(a=(n=e.next()).done)&&(i.push(n.value),!r||i.length!==r);a=!0);}catch(u){c=!0,o=u}finally{try{a||null==e.return||e.return()}finally{if(c)throw o}}return i}(t,r)||function(t,r){if(!t)return;if("string"==typeof t)return i(t,r);var e=Object.prototype.toString.call(t).slice(8,-1);"Object"===e&&t.constructor&&(e=t.constructor.name);if("Map"===e||"Set"===e)return Array.from(t);if("Arguments"===e||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(e))return i(t,r)}(t,r)||function(){throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}()}function i(t,r){(null==r||r>t.length)&&(r=t.length);for(var e=0,n=new Array(r);e<r;e++)n[e]=t[e];return n}System.register(["./index-legacy.084eaec5.js","./Paginator-legacy.56cf8554.js"],(function(t){"use strict";var e,i,a,c,u,l,s,f,h,d,p,g,y,v,m,w,b,x,k,L,$,E,S;return{setters:[function(t){e=t.d,i=t.a,a=t.b6,c=t.e,u=t.f,l=t.c7,s=t.J,f=t.W,h=t.bf,d=t.bc,p=t.m,g=t.bz,y=t.bA,v=t.B,m=t.bd,w=t.n,b=t.bx,x=t.v,k=t.as,L=t.t,$=t.a0,E=t.ap},function(t){S=t.P}],execute:function(){var _={errored:"danger",succeeded:"success",canceled:"neutral"},j=function(t){var r=e();return u(b,{get colorScheme(){var r;return null!==(r=_[t.state])&&void 0!==r?r:"info"},get children(){return r("tasks.".concat(t.state))}})},O=function(t){var b=e(),x="undone"===t.done?"cancel":"delete",k="done"===t.done&&"errored"===t.state,L=o(i((function(){return a.post("/admin/task/".concat(t.type,"/").concat(x,"?tid=").concat(t.id))})),2),$=L[0],E=L[1],S=o(i((function(){return a.post("/admin/task/".concat(t.type,"/retry?tid=").concat(t.id))})),2),_=S[0],O=S[1],I=o(c(!1),2),P=I[0],C=I[1];return u(p,{get when(){return!P()},get children(){return u(l,{get bgColor(){return s("$background","$neutral3")()},w:"$full",overflowX:"auto",shadow:"$md",rounded:"$lg",p:"$2",direction:{"@initial":"column","@xl":"row"},spacing:"$2",get children(){return[u(f,{w:"$full",alignItems:"start",spacing:"$1",get children(){return[u(h,{size:"sm",css:{wordBreak:"break-all"},get children(){return t.name}}),u(j,{get state(){return t.state}}),u(d,{css:{wordBreak:"break-all"},get children(){return t.status}}),u(p,{get when(){return t.error},get children(){return u(d,{color:"$danger9",css:{wordBreak:"break-all"},get children(){return t.error}})}}),u(g,{w:"$full",trackColor:"$info3",rounded:"$full",size:"sm",get value(){return t.progress},get children(){return u(y,{color:"$info8",rounded:"$md"})}})]}}),u(l,{direction:{"@initial":"row","@xl":"column"},justifyContent:{"@xl":"center"},spacing:"$1",get children(){return[u(p,{get when(){return t.canRetry},get children(){return u(v,{disabled:!k,display:k?"block":"none",get loading(){return _()},onClick:(t=n(r().mark((function t(){var e;return r().wrap((function(t){for(;;)switch(t.prev=t.next){case 0:return t.next=2,O();case 2:e=t.sent,m(e,(function(){w.info(b("tasks.retry")),C(!0)}));case 4:case"end":return t.stop()}}),t)}))),function(){return t.apply(this,arguments)}),get children(){return b("tasks.retry")}});var t}}),u(v,{colorScheme:"danger",get loading(){return $()},onClick:(e=n(r().mark((function t(){var e;return r().wrap((function(t){for(;;)switch(t.prev=t.next){case 0:return t.next=2,E();case 2:e=t.sent,m(e,(function(){w.success(b("global.delete_success")),C(!0)}));case 4:case"end":return t.stop()}}),t)}))),function(){return e.apply(this,arguments)}),get children(){return b("global.".concat(x))}})];var e}})]}})}})},I=function(t){var l=e(),s=o(i((function(){return a.get("/admin/task/".concat(t.type,"/").concat(t.done))})),2),d=s[0],g=s[1],y=o(c([]),2),w=y[0],b=y[1],_=function(){var t=n(r().mark((function t(){var e;return r().wrap((function(t){for(;;)switch(t.prev=t.next){case 0:return t.next=2,g();case 2:e=t.sent,m(e,(function(t){var r;return b(null!==(r=null==t?void 0:t.sort((function(t,r){return t.id>r.id?1:-1})))&&void 0!==r?r:[])}));case 4:case"end":return t.stop()}}),t)})));return function(){return t.apply(this,arguments)}}();if(_(),"undone"===t.done){var j=setInterval(_,2e3);k((function(){return clearInterval(j)}))}var I=o(i((function(){return a.post("/admin/task/".concat(t.type,"/clear_done"))})),2),P=I[0],C=I[1],A=o(i((function(){return a.post("/admin/task/".concat(t.type,"/clear_succeeded"))})),2),G=A[0],N=A[1],T=o(c(1),2),z=T[0],F=T[1],B=L((function(){var t=20*(z()-1),r=t+20;return w().slice(t,r)}));return u(f,{w:"$full",alignItems:"start",spacing:"$2",get children(){return[u(h,{size:"lg",get children(){return l("tasks.".concat(t.done))}}),u(p,{get when(){return"done"===t.done},get children(){return u($,{spacing:"$2",get children(){return[u(v,{colorScheme:"accent",get loading(){return d()},onClick:_,get children(){return l("global.refresh")}}),u(v,{get loading(){return P()},onClick:(e=n(r().mark((function t(){var e;return r().wrap((function(t){for(;;)switch(t.prev=t.next){case 0:return t.next=2,C();case 2:e=t.sent,m(e,(function(){return _()}));case 4:case"end":return t.stop()}}),t)}))),function(){return e.apply(this,arguments)}),get children(){return l("global.clear")}}),u(v,{colorScheme:"success",get loading(){return G()},onClick:(t=n(r().mark((function t(){var e;return r().wrap((function(t){for(;;)switch(t.prev=t.next){case 0:return t.next=2,N();case 2:e=t.sent,m(e,(function(){return _()}));case 4:case"end":return t.stop()}}),t)}))),function(){return t.apply(this,arguments)}),get children(){return l("tasks.clear_succeeded")}})];var t,e}})}}),u(f,{w:"$full",spacing:"$2",get children(){return u(x,{get each(){return B()},children:function(r){return u(O,E(r,t))}})}}),u(S,{get total(){return w().length},defaultPageSize:20,onChange:function(t){F(t)}})]}})};t("T",(function(t){var r=e();return u(f,{w:"$full",alignItems:"start",spacing:"$4",get children(){return[u(h,{size:"xl",get children(){return r("tasks.".concat(t.type))}}),u(f,{w:"$full",spacing:"$2",get children(){return u(x,{each:["undone","done"],children:function(r){return u(I,{get type(){return t.type},done:r,get canRetry(){return t.canRetry}})}})}})]}})}))}}}))}();
