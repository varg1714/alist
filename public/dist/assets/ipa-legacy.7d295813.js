!function(){function n(n,t){return function(n){if(Array.isArray(n))return n}(n)||function(n,e){var t=null==n?null:"undefined"!=typeof Symbol&&n[Symbol.iterator]||n["@@iterator"];if(null==t)return;var r,c,i=[],a=!0,o=!1;try{for(t=t.call(n);!(a=(r=t.next()).done)&&(i.push(r.value),!e||i.length!==e);a=!0);}catch(l){o=!0,c=l}finally{try{a||null==t.return||t.return()}finally{if(o)throw c}}return i}(n,t)||function(n,t){if(!n)return;if("string"==typeof n)return e(n,t);var r=Object.prototype.toString.call(n).slice(8,-1);"Object"===r&&n.constructor&&(r=n.constructor.name);if("Map"===r||"Set"===r)return Array.from(n);if("Arguments"===r||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(r))return e(n,t)}(n,t)||function(){throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}()}function e(n,e){(null==e||e>n.length)&&(e=n.length);for(var t=0,r=new Array(e);t<e;t++)r[t]=n[t];return r}System.register(["./index-legacy.df96d91f.js","./useUtil-legacy.c47b6e3a.js","./File-legacy.d01a71b1.js","./api-legacy.0d4b1c15.js","./icon-legacy.a49679f7.js","./index-legacy.afa56739.js","./index-legacy.422a8271.js","./Layout-legacy.f5aac4fc.js","./Markdown-legacy.4e94f452.js","./index-legacy.778f0ba8.js","./FolderTree-legacy.38ec3911.js"],(function(e){"use strict";var t,r,c,i,a,o,l,u,f,s,y;return{setters:[function(n){t=n.d,r=n.e,c=n.f,i=n.a0,a=n.B,o=n.b9,l=n.cv,u=n.o,f=n.cu},function(n){s=n.a},function(n){y=n.F},function(){},function(){},function(){},function(){},function(){},function(){},function(){},function(){}],execute:function(){e("default",(function(){var e=t(),d=n(r(!1),2),g=d[0],m=d[1],p=n(r(!1),2),h=p[0],b=p[1],j=s().currentObjLink;return c(y,{get children(){return c(i,{spacing:"$2",get children(){return[c(a,{as:"a",get href(){return"itms-services://?action=download-manifest&url="+"".concat(o,"/i/").concat(l(encodeURIComponent(u.raw_url)+"/"+f(encodeURIComponent(u.obj.name))),".plist")},onClick:function(){m(!0)},get children(){return e("home.preview.".concat(g()?"installing":"install"))}}),c(a,{as:"a",colorScheme:"primary",get href(){return"apple-magnifier://install?url="+encodeURIComponent(j(!0))},onClick:function(){b(!0)},get children(){return e("home.preview.".concat(h()?"tr-installing":"tr-install"))}})]}})}})}))}}}))}();
