//go:build ignore

#include "../headers/vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

#define TASK_COMM_SIZE 150
#define PATH_MAX 256

char __license[] SEC("license") = "Dual MIT/GPL";

struct open_event{
  u32 pid;
  u32 uid;
  u8 comm[TASK_COMM_SIZE];
  u8 filename[PATH_MAX];
  int flags;
  u64 timestamp_ns;
  long ret;
  u64 latency;
  u64 timestamp_ns_exit;
};

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 24);
} events SEC(".maps");

struct{
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, u64);
  __type(value, struct open_event);
  __uint(max_entries, 1024);
} start_events SEC(".maps");

const struct open_event *unused __attribute__((unused));

SEC("tracepoint/syscalls/sys_enter_openat")
int handle_enter_execve( struct trace_event_raw_sys_enter *ctx){
  struct open_event event = {};
  
  u64 pid_tgid = bpf_get_current_pid_tgid();

  event.pid = pid_tgid >> 32;

  event.uid = bpf_get_current_uid_gid() >>32;

  bpf_get_current_comm(&event.comm,TASK_COMM_SIZE);

  const char *filename = (const char *) ctx->args[1];

  bpf_probe_read_user_str(event.filename,sizeof(event.filename),filename);

  event.flags = (int) ctx->args[2];

  event.timestamp_ns = bpf_ktime_get_ns();

  bpf_printk("file name: %s \n" , event.filename);
  bpf_printk("comm: %s \n", event.comm);
  
  bpf_map_update_elem(&start_events,&pid_tgid,&event,BPF_ANY);
  return 0;
}

SEC("tracepoint/syscalls/sys_exit_openat")
int handle_exit_openat_tpbtf(struct trace_event_raw_sys_exit *ctx){
  
  struct open_event *event;

  u64 pid_tgid = bpf_get_current_pid_tgid();

  event = bpf_map_lookup_elem(&start_events,&pid_tgid);
  if (!event)
    return 0;

  struct open_event *final_event;

  final_event = bpf_ringbuf_reserve(&events,sizeof(struct open_event),0);

  if(!final_event) return 0;

  long ret = ctx->ret;

  final_event->pid = event->pid;
  final_event->uid = event->uid;
  bpf_probe_read_kernel_str(final_event->filename,sizeof(final_event->filename),event->filename);
  bpf_probe_read_kernel_str(final_event->comm,sizeof(final_event->comm),event->comm);
  final_event->flags = event->flags;
  final_event->timestamp_ns = event->timestamp_ns;
  final_event->ret = ret;
  u64 now = bpf_ktime_get_ns();
  final_event->timestamp_ns_exit = now;
  final_event->latency = now - event->timestamp_ns;

  bpf_printk("openat returned: %ld\n", ret);

  bpf_ringbuf_submit(final_event,0);

  bpf_map_delete_elem(&start_events,&pid_tgid);
  return 0;
}
