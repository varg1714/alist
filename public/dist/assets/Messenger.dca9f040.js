import{f as e,bf as i,ae as w,d as y,n as b,e as R,a as r,b6 as c,bw as k,as as C,W as o,v as x,ab as N,ap as T,I as W,a0 as B,B as g,h as F,aN as H,cl as L}from"./index.86895a3d.js";const z=n=>e(i,{get children(){return n.content}}),D=n=>e(w,{get src(){return n.content}}),M={string:z,image:D},V=()=>{const n=y();b.info(n("manage.messenger-tips"));const[a,l]=R(""),[d,u]=r(()=>c.post("/admin/message/get")),[p,h]=r(()=>c.post("/admin/message/send",{message:a()})),[m,S]=k([]),s=async()=>{const t=await u();F(t,I=>{S(L($=>$.push(I)))})},f=async()=>{const t=await h();H(t)},v=setInterval(s,1e3);return C(()=>clearInterval(v)),e(o,{spacing:"$2",h:"$full",alignItems:"start",get children(){return[e(o,{w:"$full",spacing:"$2",alignItems:"start",p:"$2",rounded:"$lg",border:"1px solid var(--hope-colors-neutral6)",get children(){return[e(i,{size:"xl",get children(){return n("manage.received_msgs")}}),e(x,{each:m,children:t=>e(N,T({get component(){return M[t.type]}},t))})]}}),e(W,{w:"$full",get value(){return a()},onInput:t=>l(t.currentTarget.value)}),e(B,{spacing:"$2",get children(){return[e(g,{colorScheme:"accent",get loading(){return d()},onClick:s,get children(){return n("manage.receive")}}),e(g,{get loading(){return p()},onClick:f,get children(){return n("manage.send")}})]}})]}})};export{V as Messenger,M as Shower,V as default};
