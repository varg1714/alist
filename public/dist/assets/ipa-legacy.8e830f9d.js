!function(){function n(n,t){return function(n){if(Array.isArray(n))return n}(n)||function(n,e){var t=null==n?null:"undefined"!=typeof Symbol&&n[Symbol.iterator]||n["@@iterator"];if(null==t)return;var r,i,o=[],c=!0,a=!1;try{for(t=t.call(n);!(c=(r=t.next()).done)&&(o.push(r.value),!e||o.length!==e);c=!0);}catch(l){a=!0,i=l}finally{try{c||null==t.return||t.return()}finally{if(a)throw i}}return o}(n,t)||function(n,t){if(!n)return;if("string"==typeof n)return e(n,t);var r=Object.prototype.toString.call(n).slice(8,-1);"Object"===r&&n.constructor&&(r=n.constructor.name);if("Map"===r||"Set"===r)return Array.from(n);if("Arguments"===r||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(r))return e(n,t)}(n,t)||function(){throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}()}function e(n,e){(null==e||e>n.length)&&(e=n.length);for(var t=0,r=new Array(e);t<e;t++)r[t]=n[t];return r}System.register(["./index-legacy.2142647d.js","./useUtil-legacy.e31a3074.js","./File-legacy.80ad4513.js","./api-legacy.6769b59f.js","./icon-legacy.90eed3ef.js","./index-legacy.171b15d4.js","./index-legacy.5d9d95de.js","./Layout-legacy.8559b25e.js","./Markdown-legacy.928b5cb3.js","./index-legacy.0f1bc626.js","./FolderTree-legacy.1382bbc7.js"],(function(e){"use strict";var t,r,i,o,c,a,l,u,f,s,d;return{setters:[function(n){t=n.d,r=n.e,i=n.f,o=n.a0,c=n.B,a=n.b9,l=n.cv,u=n.o,f=n.cu},function(n){s=n.a},function(n){d=n.F},function(){},function(){},function(){},function(){},function(){},function(){},function(){},function(){}],execute:function(){e("default",(function(){var e=t(),y=n(r(!1),2),g=y[0],m=y[1],b=n(r(!1),2),p=b[0],h=b[1],j=s().currentObjLink;return i(d,{get children(){return i(o,{spacing:"$2",get children(){return[i(c,{as:"a",get href(){return"itms-services://?action=download-manifest&url="+"".concat(a,"/i/").concat(l(encodeURIComponent(u.raw_url)+"/"+f(encodeURIComponent(u.obj.name))),".plist")},onClick:function(){m(!0)},get children(){return e("home.preview.".concat(g()?"installing":"install"))}}),i(c,{as:"a",colorScheme:"primary",get href(){return"apple-magnifier://install?url="+encodeURIComponent(j(!0))},onClick:function(){h(!0)},get children(){return e("home.preview.".concat(p()?"tr-installing":"tr-install"))}})]}})}})}))}}}))}();
