!function(){function e(e,t){return function(e){if(Array.isArray(e))return e}(e)||function(e,n){var t=null==e?null:"undefined"!=typeof Symbol&&e[Symbol.iterator]||e["@@iterator"];if(null==t)return;var r,i,o=[],c=!0,a=!1;try{for(t=t.call(e);!(c=(r=t.next()).done)&&(o.push(r.value),!n||o.length!==n);c=!0);}catch(l){a=!0,i=l}finally{try{c||null==t.return||t.return()}finally{if(a)throw i}}return o}(e,t)||function(e,t){if(!e)return;if("string"==typeof e)return n(e,t);var r=Object.prototype.toString.call(e).slice(8,-1);"Object"===r&&e.constructor&&(r=e.constructor.name);if("Map"===r||"Set"===r)return Array.from(e);if("Arguments"===r||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(r))return n(e,t)}(e,t)||function(){throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}()}function n(e,n){(null==n||n>e.length)&&(n=e.length);for(var t=0,r=new Array(n);t<n;t++)r[t]=e[t];return r}System.register(["./index-legacy.f69e2a17.js","./Folder-legacy.3b14f1e4.js","./Layout-legacy.29fde753.js","./useUtil-legacy.24081550.js","./index-legacy.ca9f920c.js","./icon-legacy.75cb0dea.js","./Paginator-legacy.3995bbef.js","./index-legacy.ce4378b5.js","./api-legacy.b1d3c522.js","./Markdown-legacy.031c771b.js","./index-legacy.443bd86e.js","./FolderTree-legacy.b75aec4e.js"],(function(n){"use strict";var t,r,i,o,c,a,l,u,s,g,f,d,h,m,b,y,j,p,w,x,v,A,$,S,k,z,C,M,I,E,O,P;return{setters:[function(e){t=e.f,r=e.a0,i=e.z,o=e.aE,c=e.aH,a=e.cm,l=e.m,u=e.aF,s=e.bv,g=e.ah,f=e.a7,d=e.ag,h=e.bc,m=e.by,b=e.bI,y=e.d,j=e.e,p=e.K,w=e.cn,x=e.co,v=e.cp,A=e.ap,$=e.v,S=e.o,k=e.W},function(e){z=e.b},function(e){C=e.a,M=e.M},function(e){I=e.c},function(e){E=e.n},function(e){O=e.g,P=e.O},function(){},function(){},function(){},function(){},function(){},function(){}],execute:function(){var F=[{name:"name",textAlign:"left",w:{"@initial":"76%","@md":"50%"}},{name:"size",textAlign:"right",w:{"@initial":"24%","@md":"17%"}},{name:"modified",textAlign:"right",w:{"@initial":0,"@md":"33%"}}],H=function(e){if((0,I().isHide)(e.obj))return null;var n=C().setPathAs,y=z({id:1}).show;return t(M.div,{initial:{opacity:0,scale:.95},animate:{opacity:1,scale:1},transition:{duration:.2},style:{width:"100%"},get children(){return t(r,{class:"list-item",w:"$full",p:"$2",rounded:"$lg",transition:"all 0.3s",get _hover(){return{transform:"scale(1.01)",bgColor:i()}},as:E,get href(){return e.obj.name},onMouseEnter:function(){n(e.obj.name,e.obj.is_dir,!0)},onContextMenu:function(n){o((function(){c(!1),a(e.index,!0,!0)})),y(n,{props:e.obj})},get children(){return[t(r,{class:"name-box",spacing:"$1",get w(){return F[0].w},get children(){return[t(l,{get when(){return u()},get children(){return t(s,{"on:click":function(e){e.stopPropagation()},get checked(){return e.obj.selected},onChange:function(n){a(e.index,n.target.checked)}})}}),t(g,{class:"icon",boxSize:"$6",get color(){return f()},get as(){return O(e.obj)},mr:"$1","on:click":function(n){e.obj.type===P.IMAGE&&(n.stopPropagation(),n.preventDefault(),d.emit("gallery",e.obj.name))}}),t(h,{class:"name",css:{whiteSpace:"nowrap",overflow:"hidden",textOverflow:"ellipsis"},get title(){return e.obj.name},get children(){return e.obj.name}})]}}),t(h,{class:"size",get w(){return F[1].w},get textAlign(){return F[1].textAlign},get children(){return m(e.obj.size)}}),t(h,{class:"modified",display:{"@initial":"none","@md":"inline"},get w(){return F[2].w},get textAlign(){return F[2].textAlign},get children(){return b(e.obj.modified)}})]}})}})};n("default",(function(){var n=y(),i=e(j(),2),a=i[0],g=i[1],f=e(j(!1),2),d=f[0],m=f[1];p((function(){a()&&w(a(),d())}));var b=function(e){return{fontWeight:"bold",fontSize:"$sm",color:"$neutral11",textAlign:e.textAlign,cursor:"pointer",onClick:function(){e.name===a()?m(!d()):o((function(){g(e.name),m(!1)}))}}};return t(k,{class:"list",w:"$full",spacing:"$1",get children(){return[t(r,{class:"title",w:"$full",p:"$2",get children(){return[t(r,{get w(){return F[0].w},spacing:"$1",get children(){return[t(l,{get when(){return u()},get children(){return t(s,{get checked(){return x()},get indeterminate(){return v()},onChange:function(e){c(e.target.checked)}})}}),t(h,A((function(){return b(F[0])}),{get children(){return n("home.obj.".concat(F[0].name))}}))]}}),t(h,A({get w(){return F[1].w}},(function(){return b(F[1])}),{get children(){return n("home.obj.".concat(F[1].name))}})),t(h,A({get w(){return F[2].w}},(function(){return b(F[2])}),{display:{"@initial":"none","@md":"inline"},get children(){return n("home.obj.".concat(F[2].name))}}))]}}),t($,{get each(){return S.objs},children:function(e,n){return t(H,{obj:e,get index(){return n()}})}})]}})}))}}}))}();
