/**
 * 统一响应结构
 * */

export interface  Response<T>{
    code:number,
    data:T,
    message:string
}
export function success(res:Response<any>):boolean{
    return res.code === 200;
}
