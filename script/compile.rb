#! /usr/bin/env ruby

class WorkerFile

  RE_FUNC = /^func ([a-z]+)\((.*)\) \((.*), error\) {$/i

  def initialize filename, target
    @filename = filename
    @target = target
    @funcs = {}
  end

  def compile
    @lines = File.readlines(@filename)
    insert_fmt
    @lines.grep(RE_FUNC).each &method(:inspect_func)
    replace_calls
    insert_code
    write_file
  end

  def write_file
    path = File.join @target, File.basename(@filename)
    File.open path, 'w' do |f|
      f.write @lines.join('')
    end
  end

  def inspect_func decl
    decl =~ RE_FUNC
    name = $1
    type = $3
    params = parse_params $2
    body = extract_body decl
    @funcs[name] = [params, type, body]
  end

  def extract_body decl
    start = @lines.index decl
    stop = start
    while stop < decl.size
      break if @lines[stop] == "}\n"
      stop += 1
    end
    body = @lines[start..stop]
    @lines[start..stop] = []
    body
  end

  def replace_calls
    @funcs.each do |caller,(_,_,body)|
      body.each_with_index do |line, index|
        next if index == 0
        line.scan(/([a-z]+)\((.*?)\)/i).each do |(name,inner)|
          func = @funcs[name]
          next unless func
          params = func[0].map { |p| p[0] }
          args = inner.split ', '
          if args.all? { |arg| params.include? arg }
            line.gsub! "#{name}(#{inner})", "#{name}(__router, \"#{caller}\", __args, #{params.join ', '})"
          else
            line.gsub! "#{name}(#{inner})", "#{name}(__router, \"#{caller}\", __args, #{inner})"
          end
        end
      end
    end
  end

  def insert_code
    @funcs.each do |name,(params,type,body)|
      param_str = params.map { |(name,type)| "#{name} #{type}" }.join(', ')
      param_names = params.map { |p| p[0] }.join(', ')
      param_refs = params.map { |p| "&#{p[0]}" }.join(', ')
      @lines << <<-go
func #{name}(router *Router, tgtPrefix, tgtArgs string, #{param_str}) (#{type}, error) {
  args := fmt.Sprintf(\"#{(['%v'] * params.size).join(':')}\", #{param_names})
  value, err := router.Get(\"#{name}\", args, tgtPrefix, tgtArgs)
  return value.(#{type}), err
}

func define#{name}(__router *Router) {
  __router.Define(\"#{name}\", func(__args string) (interface{}, error) {
    #{params.map {|(name, type)| "var #{name} #{type}"}.join "\t\t"}
    fmt.Sscanf(__args, \"#{(['%v'] * params.size).join(':')}\", #{param_refs})

#{body[1...-1].map {|line| "\t#{line}"}.join()}  })
}
      go
    end
  end

  def parse_params str
    parts =
    type = nil
    params = []
    str.split(', ').reverse.each do |part|
      part, type = part.split ' ' if part =~ / /
      params << [part, type]
    end
    params.reverse
  end

  def insert_fmt
    unless @lines.grep(/"fmt"/).any?
      @lines[2,0] = ["import \"fmt\"\n", "\n"]
    end
  end
end

if __FILE__ == $0
  filename = ARGV[0]
  target = ARGV[1]
  file = WorkerFile.new(filename, target)
  file.compile
end
