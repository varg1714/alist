import{as as f,d as C,f as e,I as O,m as i,B as c,a0 as m,M as b,i as k,aP as R,j as $,k as v,q as w,r as F,e as g,Y as M,az as T,a as D,_ as P,ba as V,a7 as h,ah as H,bb as j,z as p,aB as q,b1 as A,W as J,v as N,p as W,au as X,bc as Y}from"./index.c80796af.js";import{f as E}from"./index.ad571019.js";import{k as G}from"./api.11f47263.js";const x=X(),S=t=>{const[n,a]=g("/");return e(M,{class:"folder-tree-box",w:"$full",overflowX:"auto",get children(){return e(x.Provider,{get value(){var o;return{value:n,onChange:r=>{a(r),t.onChange(r)},forceRoot:(o=t.forceRoot)!=null?o:!1}},get children(){return e(_,{path:"/"})}})}})},_=t=>{const[n,a]=g([]),{value:o,onChange:r,forceRoot:l}=T(x),[y,B]=D(()=>G(t.path,P(),l)),z=async()=>{if(n().length>0)return;const u=await B();Y(u,a)},{isOpen:s,onToggle:I}=f(),d=()=>o()===t.path;return e(M,{get children(){return[e(m,{spacing:"$2",get children(){return[e(i,{get when(){return!y()},get fallback(){return e(V,{size:"sm",get color(){return h()}})},get children(){return e(H,{get color(){return h()},as:E,get transform(){return s()?"rotate(90deg)":"none"},transition:"transform 0.2s",cursor:"pointer",onClick:()=>{I(),s()&&z()}})}}),e(j,{css:{whiteSpace:"nowrap"},fontSize:"$md",cursor:"pointer",px:"$1",rounded:"$md",get bgColor(){return d()?"$info8":"transparent"},get _hover(){return{bgColor:d()?"$info8":p()}},onClick:()=>{r(t.path)},get children(){return q(()=>t.path==="/",!0)()?"root":A(t.path)}})]}}),e(i,{get when(){return s()},get children(){return e(J,{mt:"$1",pl:"$4",alignItems:"start",spacing:"$1",get children(){return e(N,{get each(){return n()},children:u=>e(_,{get path(){return W(t.path,u.name)}})})}})}})]}})},U=t=>{var r;const n=C(),[a,o]=g((r=t.defaultValue)!=null?r:"");return e(F,{size:"xl",blockScrollOnMount:!1,get opened(){return t.opened},get onClose(){return t.onClose},get children(){return[e(b,{}),e(k,{get children(){return[e($,{get children(){return n("home.toolbar.choose_dst_folder")}}),e(v,{get children(){return e(S,{onChange:o})}}),e(w,{display:"flex",gap:"$2",get children(){return[e(c,{get onClick(){return t.onClose},colorScheme:"neutral",get children(){return n("global.cancel")}}),e(c,{get loading(){return t.loading},onClick:()=>{var l;return(l=t.onSubmit)==null?void 0:l.call(t,a())},get children(){return n("global.ok")}})]}})]}})]}})},Z=t=>{const{isOpen:n,onOpen:a,onClose:o}=f(),r=C();return[e(m,{w:"$full",spacing:"$2",get children(){return[e(O,{get id(){return t.id},get value(){return t.value},onInput:l=>t.onChange(l.currentTarget.value),get readOnly(){return t.onlyFolder},get onClick(){return t.onlyFolder?a:()=>{}},get placeholder(){return r(`global.${t.onlyFolder?"choose_folder":"choose_or_input_path"}`)}}),e(i,{get when(){return!t.onlyFolder},get children(){return e(c,{onClick:a,get children(){return r("global.choose")}})}})]}}),e(F,{size:"xl",get opened(){return n()},onClose:o,get children(){return[e(b,{}),e(k,{get children(){return[e(R,{}),e($,{get children(){return r("global.choose_folder")}}),e(v,{get children(){return e(S,{forceRoot:!0,get onChange(){return t.onChange}})}}),e(w,{get children(){return e(c,{onClick:o,get children(){return r("global.confirm")}})}})]}})]}})]};export{Z as F,U as M};
