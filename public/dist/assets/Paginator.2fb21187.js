import{ap as x,bw as z,t as h,f as t,a0 as C,m as g,B as l,a6 as s,v as m}from"./index.7e2ddf5f.js";import{k as w,l as $}from"./index.1c1c9d8f.js";const _=S=>{const r=x({maxShowPage:4,defaultPageSize:30,defaultCurrent:1,hideOnSinglePage:!0},S),[n,d]=z({pageSize:r.defaultPageSize,current:r.defaultCurrent}),a=h(()=>Math.ceil(r.total/n.pageSize)),P=h(()=>{const e=n.current,c=Math.max(2,e-Math.floor(r.maxShowPage/2));return Array.from({length:e-c},(p,u)=>c+u)}),f=h(()=>{const e=n.current,c=Math.min(a()-1,e+Math.floor(r.maxShowPage/2));return Array.from({length:c-e},(p,u)=>e+1+u)}),o={"@initial":"sm","@md":"md"},i=e=>{var c;d("current",e),(c=r.onChange)==null||c.call(r,e)};return t(g,{get when(){return!r.hideOnSinglePage||a()>1},get children(){return t(C,{spacing:"$1",get children(){return[t(g,{get when(){return n.current!==1},get children(){return[t(l,{size:o,get colorScheme(){return r.colorScheme},onClick:()=>{i(1)},px:"$3",children:"1"}),t(s,{size:o,get icon(){return t(w,{})},"aria-label":"Previous",get colorScheme(){return r.colorScheme},onClick:()=>{i(n.current-1)},w:"2rem !important"})]}}),t(m,{get each(){return P()},children:e=>t(l,{size:o,get colorScheme(){return r.colorScheme},onClick:()=>{i(e)},px:e>10?"$2_5":"$3",children:e})}),t(l,{size:o,get colorScheme(){return r.colorScheme},variant:"solid",get px(){return n.current>10?"$2_5":"$3"},get children(){return n.current}}),t(m,{get each(){return f()},children:e=>t(l,{size:o,get colorScheme(){return r.colorScheme},onClick:()=>{i(e)},px:e>10?"$2_5":"$3",children:e})}),t(g,{get when(){return n.current!==a()},get children(){return[t(s,{size:o,get icon(){return t($,{})},"aria-label":"Next",get colorScheme(){return r.colorScheme},onClick:()=>{i(n.current+1)},w:"2rem !important"}),t(l,{size:o,get colorScheme(){return r.colorScheme},onClick:()=>{i(a())},get px(){return a()>10?"$2_5":"$3"},get children(){return a()}})]}})]}})}})};export{_ as P};
