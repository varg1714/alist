!function(){function e(e,r,n){return r in e?Object.defineProperty(e,r,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[r]=n,e}System.register(["./index-legacy.084eaec5.js","./index-legacy.a16bddb8.js","./item_type-legacy.d1f1c701.js","./index-legacy.21abb1d1.js"],(function(r){"use strict";var n,t,u,l,i,c,a,o,g,d,f,h,b,v,y,s,p,m,O,k,w,C,j,x,D,T,E,R;return{setters:[function(e){n=e.d,t=e.f,u=e.m,l=e.a_,i=e.ai,c=e.ah,a=e.R,o=e.ak,g=e.U,d=e.I,f=e.bR,h=e.T,b=e.bS,v=e.bT,y=e.bU,s=e.bV,p=e.bW,m=e.bX,O=e.bY,k=e.v,w=e.bZ,C=e.b_,j=e.b$,x=e.F,D=e.b0},function(e){T=e.F},function(e){E=e.T},function(e){R=e.m}],execute:function(){r("I",(function(r){var S=n();return t(D,{get w(){var e;return null!==(e=r.w)&&void 0!==e?e:"100%"},display:"flex",flexDirection:"column",get children(){return[t(u,{get when(){return!r.hideLabel},get children(){var n,a;return t(l,(n={},"for",(a={}).for=a.for||{},a.for.get=function(){return r.key},e(n,"display","flex"),e(n,"alignItems","center"),"children",a.children=a.children||{},a.children.get=function(){return[i((function(){return S("settings.".concat(r.key))})),t(u,{get when(){return r.flag===T.DEPRECATED},get children(){return t(c,{ml:"$2",as:R,boxSize:"$5",color:"$danger9",verticalAlign:"middle",cursor:"pointer",onClick:function(){var e;null===(e=r.onDelete)||void 0===e||e.call(r)}})}})]},function(e,r){for(var n in r)(l=r[n]).configurable=l.enumerable=!0,"value"in l&&(l.writable=!0),Object.defineProperty(e,n,l);if(Object.getOwnPropertySymbols)for(var t=Object.getOwnPropertySymbols(r),u=0;u<t.length;u++){var l,i=t[u];(l=r[i]).configurable=l.enumerable=!0,"value"in l&&(l.writable=!0),Object.defineProperty(e,i,l)}}(n,a),n))}}),t(a,{get fallback(){return t(o,{get children(){return S("settings_other.unknown_type")}})},get children(){return[t(g,{get when(){return[E.String,E.Number].includes(r.type)},get children(){return t(d,{get type(){return r.type===E.Number?"number":""},get id(){return r.key},get value(){return r.value},onInput:function(e){var n;return null===(n=r.onChange)||void 0===n?void 0:n.call(r,e.currentTarget.value)},get readOnly(){return r.flag===T.READONLY}})}}),t(g,{get when(){return r.type===E.Bool},get children(){return t(f,{get id(){return r.key},get defaultChecked(){return"true"===r.value},onChange:function(e){var n;return null===(n=r.onChange)||void 0===n?void 0:n.call(r,e.currentTarget.checked?"true":"false")},get readOnly(){return r.flag===T.READONLY}})}}),t(g,{get when(){return r.type===E.Text},get children(){return t(h,{get id(){return r.key},get value(){return r.value},onChange:function(e){var n;return null===(n=r.onChange)||void 0===n?void 0:n.call(r,e.currentTarget.value)},get readOnly(){return r.flag===T.READONLY}})}}),t(g,{get when(){return r.type===E.Select},get children(){return t(b,{get id(){return r.key},get defaultValue(){return r.value},onChange:function(e){var n;return null===(n=r.onChange)||void 0===n?void 0:n.call(r,e)},get readOnly(){return r.flag===T.READONLY},get children(){return[t(v,{get children(){return[t(y,{get children(){return S("global.choose")}}),t(s,{}),t(p,{})]}}),t(m,{get children(){return t(O,{get children(){return t(k,{get each(){var e;return null===(e=r.options)||void 0===e?void 0:e.split(",")},children:function(e){return t(w,{value:e,get children(){return[t(C,{get children(){return S("settings.".concat(r.key,"s.").concat(e))}}),t(j,{})]}})}})}})}})]}})}})]}}),t(x,{get children(){return i((function(){return!!r.help}),!0)()?S("settings.".concat(r.key,"-tips")):""}})]}})}))}}}))}();
