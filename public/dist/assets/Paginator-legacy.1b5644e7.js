!function(){function e(e,n){return function(e){if(Array.isArray(e))return e}(e)||function(e,r){var n=null==e?null:"undefined"!=typeof Symbol&&e[Symbol.iterator]||e["@@iterator"];if(null==n)return;var t,o,c=[],u=!0,i=!1;try{for(n=n.call(e);!(u=(t=n.next()).done)&&(c.push(t.value),!r||c.length!==r);u=!0);}catch(l){i=!0,o=l}finally{try{u||null==n.return||n.return()}finally{if(i)throw o}}return c}(e,n)||function(e,n){if(!e)return;if("string"==typeof e)return r(e,n);var t=Object.prototype.toString.call(e).slice(8,-1);"Object"===t&&e.constructor&&(t=e.constructor.name);if("Map"===t||"Set"===t)return Array.from(e);if("Arguments"===t||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(t))return r(e,n)}(e,n)||function(){throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}()}function r(e,r){(null==r||r>e.length)&&(r=e.length);for(var n=0,t=new Array(r);n<r;n++)t[n]=e[n];return t}System.register(["./index-legacy.f2f6d3e1.js","./index-legacy.297ddf3e.js"],(function(r){"use strict";var n,t,o,c,u,i,l,a,f,h,g;return{setters:[function(e){n=e.ap,t=e.bw,o=e.t,c=e.f,u=e.a0,i=e.m,l=e.B,a=e.a6,f=e.v},function(e){h=e.j,g=e.k}],execute:function(){r("P",(function(r){var m=n({maxShowPage:4,defaultPageSize:30,defaultCurrent:1,hideOnSinglePage:!0},r),d=e(t({pageSize:m.defaultPageSize,current:m.defaultCurrent}),2),s=d[0],S=d[1],y=o((function(){return Math.ceil(m.total/s.pageSize)})),p=o((function(){var e=s.current,r=Math.max(2,e-Math.floor(m.maxShowPage/2));return Array.from({length:e-r},(function(e,n){return r+n}))})),v=o((function(){var e=s.current,r=Math.min(y()-1,e+Math.floor(m.maxShowPage/2));return Array.from({length:r-e},(function(r,n){return e+1+n}))})),x={"@initial":"sm","@md":"md"},w=function(e){var r;S("current",e),null===(r=m.onChange)||void 0===r||r.call(m,e)};return c(i,{get when(){return!m.hideOnSinglePage||y()>1},get children(){return c(u,{spacing:"$1",get children(){return[c(i,{get when(){return 1!==s.current},get children(){return[c(l,{size:x,get colorScheme(){return m.colorScheme},onClick:function(){w(1)},px:"$3",children:"1"}),c(a,{size:x,get icon(){return c(h,{})},"aria-label":"Previous",get colorScheme(){return m.colorScheme},onClick:function(){w(s.current-1)},w:"2rem !important"})]}}),c(f,{get each(){return p()},children:function(e){return c(l,{size:x,get colorScheme(){return m.colorScheme},onClick:function(){w(e)},px:e>10?"$2_5":"$3",children:e})}}),c(l,{size:x,get colorScheme(){return m.colorScheme},variant:"solid",get px(){return s.current>10?"$2_5":"$3"},get children(){return s.current}}),c(f,{get each(){return v()},children:function(e){return c(l,{size:x,get colorScheme(){return m.colorScheme},onClick:function(){w(e)},px:e>10?"$2_5":"$3",children:e})}}),c(i,{get when(){return s.current!==y()},get children(){return[c(a,{size:x,get icon(){return c(g,{})},"aria-label":"Next",get colorScheme(){return m.colorScheme},onClick:function(){w(s.current+1)},w:"2rem !important"}),c(l,{size:x,get colorScheme(){return m.colorScheme},onClick:function(){w(y())},get px(){return y()>10?"$2_5":"$3"},get children(){return y()}})]}})]}})}})}))}}}))}();
