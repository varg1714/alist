!function(){function n(n,r){return function(n){if(Array.isArray(n))return n}(n)||function(n,t){var r=null==n?null:"undefined"!=typeof Symbol&&n[Symbol.iterator]||n["@@iterator"];if(null==r)return;var e,o,u=[],i=!0,a=!1;try{for(r=r.call(n);!(i=(e=r.next()).done)&&(u.push(e.value),!t||u.length!==t);i=!0);}catch(l){a=!0,o=l}finally{try{i||null==r.return||r.return()}finally{if(a)throw o}}return u}(n,r)||function(n,r){if(!n)return;if("string"==typeof n)return t(n,r);var e=Object.prototype.toString.call(n).slice(8,-1);"Object"===e&&n.constructor&&(e=n.constructor.name);if("Map"===e||"Set"===e)return Array.from(n);if("Arguments"===e||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(e))return t(n,r)}(n,r)||function(){throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}()}function t(n,t){(null==t||t>n.length)&&(t=n.length);for(var r=0,e=new Array(t);r<t;r++)e[r]=n[r];return e}System.register(["./index-legacy.4335ae98.js","./useUtil-legacy.3402c62f.js","./Markdown-legacy.9b42dbd1.js","./api-legacy.3605fdb1.js"],(function(t){"use strict";var r,e,o,u,i,a;return{setters:[function(n){r=n.f,e=n.Z,o=n.o,u=n.bG},function(n){i=n.d},function(n){a=n.M},function(){}],execute:function(){t("default",(function(){var t=n(i(),1)[0];return r(e,{get loading(){return t.loading},get children(){return r(a,{get children(){var n,r;return function(n){return o.obj.name.endsWith(".md")?n:"```"+u(o.obj.name)+"\n"+n+"\n```"}(null!==(n=null===(r=t())||void 0===r?void 0:r.content)&&void 0!==n?n:"")}})}})}))}}}))}();
