#ifndef MSQUIC_CONN_H
#define MSQUIC_CONN_H

#include <msquic.h>

// 封装的连接结构
typedef struct {
    QUIC_CONNECTION* conn;
} MsQuicConn;

// 初始化连接
int MsQuicConnInit(MsQuicConn* conn);

// 连接到远程主机
int MsQuicConnConnect(MsQuicConn* conn, const char* host, int port);

// 发送数据
int MsQuicConnSend(MsQuicConn* conn, const char* data, size_t len);

// 接收数据
int MsQuicConnReceive(MsQuicConn* conn, char* buffer, size_t len);

// 关闭连接
void MsQuicConnClose(MsQuicConn* conn);

#endif